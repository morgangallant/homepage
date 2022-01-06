import * as React from "react";
import Head from "next/head";
import Layout, { siteTitle } from "../components/layout";
import utilStyles from "../styles/utils.module.css";
import { getSortedPostsData } from "../lib/posts";
import Link from "next/link";
import Date from "../components/date";
import { GetStaticProps } from "next";

type searchResults = {
  search: string;
  project: string;
  query: string;
  filter: string;
  brrr: boolean;
  duration_ms: number;
  documents: {
    document: string;
    title: string;
    url: string;
    samples: string[];
  }[];
};

export default function Home({
  allPostsData,
}: {
  allPostsData: {
    date: string;
    title: string;
    id: string;
  }[];
}) {
  // Search Functionality
  const [query, setQuery] = React.useState("");
  const [results, setResults] = React.useState<searchResults | null>(null);
  React.useEffect(() => {
    if (query == "") {
      setResults(null);
      return;
    }
    const delayed = setTimeout(async () => {
      const res = await fetch(`/api/search?q=${query}`);
      const data = await res.json();
      setResults(data);
    }, 250);
    return () => clearTimeout(delayed);
  }, [query]);

  // Main Component
  return (
    <Layout home>
      <Head>
        <title>{siteTitle}</title>
      </Head>
      <section className={utilStyles.headingMd}>
        <p className={utilStyles.centered}>
          Founder, programmer, optimist. Currently working on{" "}
          <a href="https://operand.ai">Operand</a>.
        </p>
        <p className={utilStyles.centered}>
          Need to get in touch? Send me an{" "}
          <a href="mailto:morgan@morgangallant.com">email</a>.
        </p>
      </section>
      <section className={`${utilStyles.headingMd} ${utilStyles.padding1px}`}>
        <div className={utilStyles.headerFlex}>
          <div>
            <h2 className={utilStyles.headingLg}>Blog</h2>
          </div>
          <div>
            <form>
              <input
                className={utilStyles.searchBar}
                type="text"
                autoComplete="false"
                placeholder="Search"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
              />
            </form>
          </div>
        </div>
        <ul className={utilStyles.list}>
          {results
            ? results.documents.map((doc) => (
                <li className={utilStyles.listItem} key={doc.document}>
                  <Link href={doc.url}>
                    <a>{doc.title}</a>
                  </Link>
                  <br></br>
                  <small className={utilStyles.lightText}>
                    {doc.samples.join(", ")}
                  </small>
                </li>
              ))
            : allPostsData.map(({ id, date, title }) => (
                <li className={utilStyles.listItem} key={id}>
                  <Link href={`/posts/${id}`}>
                    <a>{title}</a>
                  </Link>
                  <br />
                  <small className={utilStyles.lightText}>
                    <Date dateString={date} />
                  </small>
                </li>
              ))}
        </ul>
      </section>
      <section className={utilStyles.paddding1px}>
        <p className={utilStyles.centered}></p>
      </section>
    </Layout>
  );
}

export const getStaticProps: GetStaticProps = async () => {
  const allPostsData = getSortedPostsData();
  return {
    props: {
      allPostsData,
    },
  };
};
