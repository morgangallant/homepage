---
title: "Ethical Usage of Natural Language Tools"
date: "2020-08-05"
---

Anyone who has gotten beta access to [OpenAI’s API](https://openai.com/blog/openai-api/), or has used any of the [fantastic tools](https://gpt3examples.com/) already created with GPT-3, knows that we’ve hit a major breakthrough in natural language understanding and generation. I would be remiss not to mention the famous Stan Lee quote saying that “with great power comes great responsibility” — now that we have powerful natural language tools, we need to make sure that they are used in ways that create a positive user experience and benefit society as a whole. The pressing question at the moment is where do we draw the so-called “line in the sand”, better put, what criterion do we use to judge whether or not an application is an ethical application of a natural language tool? Generally speaking, these types of ethical dilemmas are really difficult to solve (would you kill one person to save five?), so if you’re reading this, definitely don’t expect a concrete answer as to whether or not an arbitrary natural language tool is ethical or not. Rather, this essay will take a principled approach when thinking about these issues. It’s also worthwhile to mention that I think I’m uniquely qualified to write about this issue, given that I’ve built and released apps that have tested these ethical boundaries.

I got access to the GPT-3 API on July 13th, 2020 and released my first app using it by July 17th - a “complaint department” which responds to user complaints in the personality of Sergeant Gunnery Hartman from the movie Full Metal Jacket. I was (and still am) fascinated by the idea that it’s possible for GPT-3 to take on different personalities and emulate the behaviour of others, even fictional characters. Throughout my childhood, my father made sure that I had a great education of 80s/90s war movies, and one of my favorite characters from those movies is the drill sergeant from Full Metal Jacket, known for his clever, yet not so politically correct one-liners. I released this application with good intention, to give people a laugh and maybe even motivate them in some dark and twisted way to achieve their goals. However, at the time, I was certainly ignorant to a lot of the potentially negative impacts a tool like this can have. Even though the site had multiple warnings about how the generated content could be vulgar and offensive, even I had trouble not taking some of the responses to heart.

The site launched at 7:08PM on a friday, and by 7:39PM I had received a message from the Greg Brockman, CTO of OpenAI, asking for me to take it down temporarily and schedule a meeting with the team. In the time that it was online (about 40 minutes), the site got ~2.6k requests and I had received over 50 screenshots from people sharing notable responses. After a quick discussion with Greg and the team, it was decided that the site should remain offline for the foreseeable future, an action which I wholeheartedly agreed with. This was a great example of something that shouldn’t have been built with GPT-3, and as Greg put it, a developer should be ultimately responsible and ready to condone any outputs from their tools.

I had a video call with the OpenAI team the next week to talk about something new I was working on, and was extremely impressed at how the team is ensuring that GPT-3 is used ethically. Of course, they’re still working on developing their internal criterion for approving GPT-3 applications, but in general they’re taking a proactive stance on making sure that the technology is being used for good. After this meeting, I got approval to launch another tool which this time, hopefully did some good.

I launched [Regex is Hard](https://regexishard.com/) on July 31st, got a lot of positive feedback and got ~11.7k requests in the first 36 hours of launch. As the title suggests, Regex isn’t very fun to work with since it is a language designed to be parsed by computers. It was surprisingly easy (just 2 examples) to get GPT-3 to generate fairly accurate regular expressions from plain english, and can even (sometimes) go the other way. I think this is an example of a universally good use case for natural language tools, creating useful tools for others that help with a mundane or repetitive task.

I’m a big believer that every decision should be made through the lens of a set of strong personal moral principles that ultimately define who you are. Personally, I tend to take a utilitarian approach to solving most problems, except when doing so would be unethical. For a concrete example, I don’t believe that should’ve released the complaint department tool I wrote. Although it was intended to be comedic, it was hard to ignore just how much negativity was output from the tool, and how that negativity could harm the mental health of my users. It’s one thing to write a potentially harmful comment or joke, it’s another thing to tailor that towards a specific user using something as powerful as GPT-3.

Natural language applications are much more intimate with their users than other applications, because typically the user is able to have a conversation with the tool rather than just pressing buttons on a user interface. A great example of this which I do believe to be an excellent use of natural language technology is [Inwords](https://www.inwords.ai/), a platform providing affordable, automated therapy sessions to “normalize and provide access to mental wellness” for its users. Technologies like Inwords work to use the power of natural language for good, and to develop close, meaningful relationships with users.

As I work to build new and more powerful natural language tools, I developed a set of four questions that I ask myself every time I’m about to release something new:

does the tool abuse any potentially intimate connections with its users?
does the tool provide a universal benefit to each and every user in some way?
if you, the developer, were asked to manually serve user requests without using any automated natural language tool, would you be comfortable doing so?
can the tool be misused to unintentionally have a negative impact on users?
If you can’t stand behind your answers to any of these four questions, it may be worth considering whether or not the tool should be built/released in the first place. I’m especially critical of the fourth question (potential for misuse) due to the relatively young age of powerful natural language tools — it seems we don’t have a complete understanding of the failure modes of these large models and the impact that it could have on users.

With the release of GPT-3, I’m fully expecting the next 1-2 years to be full of exciting new applications leveraging natural language to create powerful features. I’m looking forward to the day when I can [write code by just describing my intention](https://debuild.co/), or [ask Elon Musk to teach me about rockets](https://learnfromanyone.com/). I’m also really excited to be building and eventually releasing some tools of my own in the virtual assistant space. As we (the next generation of makers) set out to build these tools, it’s important for us to set a strong precedent of both building universally great tools with this technology, and being proactive in ensuring that the apps we build aren’t misused.