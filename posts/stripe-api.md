---
title: "Beating the Crap out of the Stripe API (Respectfully)"
date: "2021-11-22"
---

Alternate Title: Fetching large volumes of data concurrently from a cursor-paginated API.

This past weekend was spent largely on one specific problem - how can we fetch huge volumes of data from the Stripe API as quickly as possible? To frame the problem a little better, it's probably worth giving some background. Our initial implementation was the simplest possible one (as it should be), which was roughly the following:

```
let subscriptions = [];
let more = true;
while (more) {
  let url = `https://api.stripe.com/v1/subscriptions?limit=100${
    subscriptions.length > 0 ? `&starting_after=${subscriptions[subscriptions.length - 1].id}` : ""
  }`;
  const response = await fetch(url, {
    method: "GET",
    headers: {
      Authorization: "Bearer " + operand.env("STRIPE_SECRET_KEY"),
    },
  });
  // omitted: error checking and other nonsense
  const json = await response.json();
  more = json.has_more;
  subscriptions = active.concat(json.data);
}
```

Essentially, we'd fetch the first 100 subscriptions, then if there were more, we'd fetch the next 100, until there are no more subscriptions left. This implementation worked fantastic in development, where we were using a test mode Stripe account with a total of, wait for it, 4 customers. I'm not sure what we were on that day, but we shipped this into production and onboarded one of the first external users onto the system. We learned of our mistake pretty quickly, as this specific user we onboarded ran a large business w/ ~60k customers per month (note: this is our conservative estimate based on the company's published stats). The above implementation is still technically correct for this volume of customers, yet we run into issues with the runtime speed. Assuming ~60k items, fetching 100 (the maximum allowed by Stripe) at a time taking ~300ms per request, we'd expect the fetch to take around 180 seconds (and it did, yet our HTTP requests timed out before we returned a result to the user).

Alright, 180 seconds is our baseline - how can we make it faster? We want this system to operate in largely realtime; users shouldn't have to wait for answers. Luckily, Stripe itself has pretty generous rate limits, 25 (read) requests per second in test mode and 100 in live mode. This means that there is ample room for us to speed this up considerably via concurrent requests.

When you hear the word "concurrency", what do you think of? Personally, I definitely don't think of JavaScript, rather, I think of Go. It's a language purpose built for this kind of stuff. I'm really sorry if you saw that JS code snippet earlier and got excited about some beautiful JS implementation, as there is something I failed to mention. That JS code is running inside a V8 runtime using the [v8go package](https://github.com/rogchap/v8go), and using the magic of polyfills, we can do our implementation of this in Go and call the function from JS. With this approach, our JS code becomes the following:

```
let subscriptions = stripe.fetch({
endpoint: "https://api.stripe.com/v1/subscriptions",
});
```

And yes, I am sorry, but that will be the last JavaScript code snippet in this entire blog article. We're operating in Go(land) now, meaning we get to use all the concurrency primitives of the language in addition to a bunch of useful open source packages.

The typical way of making these types of large-volume API requests concurrent is to use the (semi-standard) `limit` and `offset` parameters. This implementation works because you can spin up n workers to request blocks of 100n items at a time. Essentially, this means that even though Stripe limits you to fetching 100 items at a time per API request, you can make a bunch of API requests simultaneously to fetch a large range of items. Once you notice that you're not getting any more items back from Stripe, you're done and you can return the data to the caller.

The problem is that Stripe doesn't support the offset parameter in their API requests. They used to, but they switched over to using `starting_after` and `ending_before` parameters in 2014 (aka cursor-based pagination). This makes a lot of sense from a database perspective, as offsets are rather expensive (and doesn't scale well) whereas starting a scan from a particular (indexed) ID string is super cheap. For backwards compatibility reasons, although the parameter itself is depreciated, most (about 90%) of their main endpoints still support it (though this is largely undocumented). We weren't comfortable relying on a depreciated feature of their API long-term, so we had to figure out a way to fetch the data concurrently with their cursor-based pagination system.

This is the documentation for the `starting_after` parameter:

```
starting_after

A cursor for use in pagination. starting_after is an object ID that defines your place in the list. For instance, if you make a list request and receive 100 objects, ending with obj_foo, your subsequent call can include starting_after=obj_foo in order to fetch the next page of the list.
```

Wait a minute, we need to pass in the ID of the object? That's bad right? The goal here is to fetch data concurrently, yet how the heck can we fire off multiple requests at once if we need data from the previous request to create the next? We banged our head against the wall for a couple hours on this, went through most of the well-document stages of grief, and then in an act of desperation we started scouring the internet for ideas. Someone has definitely had this problem before right? Yes, yes they have. We found this really well written [blog post](https://brandur.org/fragments/offset-pagination) by [@brandur](https://twitter.com/brandur) explaining some details about fetching data in parallel from cursor-based APIs. In his post, he suggested slicing the overall time range into pieces and fetching those pieces independently. He even mentioned the Stripe API specifically, the fact that they support a `created` parameter on all of their API requests:

```
created

A filter on the list based on the object created field. The value can be a string with an integer Unix timestamp, or it can be a dictionary with the following options:

created.gt
Return results where the created field is greater than this value.

created.gte
Return results where the created field is greater than or equal to this value.

created.lt
Return results where the created field is less than this value.

created.lte
Return results where the created field is less than or equal to this value.
```

This was exactly what we needed - we could split the overall search space into n pieces and fetch data from those pieces concurrently, using the `starting_after` parameter inside the pieces themselves to fetch all the data. In theory, this works great, yet a few concerns do arise (also partially mentioned in the blog post) which require some special attention:

Initial time ranges. We have no way of knowing a) how many items there are, or b) when the first item was created, so how do we choose our initial bounds for our requests?
Distribution of items. In a perfect world, data is distributed uniformly throughout a given time range. This isn't necessarily the case for Stripe data.
To explain this further, let's say you're fetching customer data from 2021. You split the time range into 12 equal segments, one per month, and fetch them all in parallel. It's possible that you had only a few new customers in most months, but in October, you had a blog post go viral and you got tens of thousands of new customers. In this case, the data itself is unevenly distributed, and the time slice for October will be stuck fetching all the data for that month in essentially the same manner as our initial implementation (single worker, 100 at a time). Put more formally, the runtime speed of the basic time slicing approach is limited by the biggest (& therefore slowest) time slice.

Since we had no idea of the distribution of the data, it was clear that we were missing something from the time range approach. Namely, we quickly realized that although we don't know the distribution initially, we do slowly learn about it as we do more and more requests.

Wait, pause, you said the word "learn" in a technical blog post? Machine learning? Training neural nets? Deploy on TPUs ([plug](https://banana.dev/))? No, no and no. I guess you could do that, but really, "learning" in this case means that we can reallocate workers from finished time slots onto those which require more attention.

Back to our fetching 2021 customer data example, where October has tens of thousands of customers and the other months have none. We initially take the naive approach, splitting the time range into 12 pieces and doing requests for all of them. When we get our results, we notice that 11 of the 12 months are empty, whereas the 12th month (October) returned a response with 100 items and a flag which tells us there are more requests. We now have 12 free workers, and the knowledge that we still have more items to fetch from October. At this point, we can take the time slice for October, slice it up once again, and repeat the same procedure (this time, each worker is processing slices of ~2.58 days). This gives us a way to continually do concurrent requests for time ranges that we know are important and contain information that we need to fetch.

This seems great, though there are some important considerations that should be mentioned, especially if you the reader are planning on implementing a version of this for yourself (it's fun!):

- Choosing the subdivision parameters is important. For us, a value of 6 seemed to work well, though we haven't experimented with it too much. Essentially, this means that when we see a time range that has more than 100 items in it, we split that time range up into 6 pieces and queue up fetches for each of those time ranges.

- After finishing the first request of a given time range denoting that there are more elements inside it to fetch, you don't need to slice up the whole time range, only the time range remaining after that initial fetch. For example, if we fetched the first 100 elements from October, rather then slicing up the entire month of October, we can look at the last fetched element for its created date and fetch from there to the end of the month, which means we aren't refetching the same data twice.

- Depending on the distribution of data itself, it will be useful to have a parameter denoting the minimum size for a time range. Rather than further subdividing these ranges, we use the simple `starting_after` fetching technique. This lowers the overall number of requests required drastically because it prevents separate individual requests for tiny subranges, likely containing less than 100 items.

Here's some more code for your viewing pleasure (the variable `ts` is the timestamp of the last fetched element for a time range we're trying to slice):

```
// Since `gte` defines the bottom of the range, we selectively lower the
// top of the range using the `lt` parameter. If div is too small, we've
// hit the limits of our subdividing and we do bigger requests to ensure
// we're saturating stripeDatumsPerRequest (100).
div := (ts - elem.end) / stripeSubdivideRangeParam
if div <= stripeDivisionMinimumSeconds {
queue = append(queue, velocityStripeTimeRange{
start: elem.start,
end: elem.end,
after: id,
})
return nil
}
for i := 0; i < stripeSubdivideRangeParam; i++ {
tr := velocityStripeTimeRange{
start: ts - div*int64(i),
end: ts - div*int64(i+1),
}
// Little bit of nonsense to deal with gte/lt shenanigans.
if i == 0 {
tr.after = id // Important - don't want to fetch duplicate elements.
tr.start += 1
} else if i == stripeSubdivideRangeParam-1 {
tr.end -= 1
}
queue = append(queue, tr)
}
```

Another consideration is rate limiting - since you are firing a bunch of concurrent requests, you gotta make sure you aren't getting 429 status codes from the API. There's an excellent [Go package](https://golang.org/x/time/rate) for this exact thing.

```
// The actual rate limits for the Stripe API in test/live mode are
// 25/100 (read) requests per second. To be safe, we use only 80%.
var limiter \*rateLimiter
if strings.HasPrefix(skey, "sk*test*") {
limiter = &rateLimiter{
Limiter: rate.NewLimiter(rate.Limit(20), 1),
used: time.Now(),
}
} else if strings.HasPrefix(skey, "sk*live*") {
limiter = &rateLimiter{
Limiter: rate.NewLimiter(rate.Limit(80), 1),
used: time.Now(),
}
}

...

limiter.Wait(params.ctx)
results, err := doStripeRequest(params.ctx, url, params.skey)

...
```

It's also an interesting question to consider how to pick initial time ranges for requests. Our initial thought was to do it between unix time 0 and now, but this seems intuitively like a bad idea. The question then becomes, is there a date where we can be sure that no production Stripe data exists before it? Of course, it's the founding date of Stripe itself:

```
// The founding date of Stripe is 4/13/2009.
// Theoretically, no Stripe data should exist before this date.
var stripeFoundingUnix = time.Date(2009, 4, 13, 0, 0, 0, 0, time.UTC).Unix()
```

In test mode (with rate limits of 20 read requests per second) running on my laptop, we managed to fetch ~40k items from the Stripe API in 41s (which works out to an average ~9.75 requests per second, about 50% of the rate limit). We suspect this poor efficiency in our testing is likely due to the distribution of data, as it was created via a script and many of the Stripe data items were 'created' at the same time, making it challenging for our time-range fetching implementation.

This blog post will be updated with experimental data from live mode when we have been able to collect it. Our suspicion is that our efficiency (as a % of rate limit) will increase when working with perhaps more realistic data.

You can use a demo of this system (on a small amount of test data) [here](https://velocity.operand.ai/modules/stripe), and you can [follow me on Twitter](https://twitter.com/morgallant) for updates. Thanks for reading!

P.S. If you're reading this and your name is Patrick Collison, please respond to my cold email 🥺.
