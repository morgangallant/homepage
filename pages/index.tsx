import * as React from "react";
import Head from "next/head";
import Layout, { siteTitle } from "../components/layout";
import utilStyles from "../styles/utils.module.css";
import { getSortedPostsData } from "../lib/posts";
import Link from "next/link";
import Date from "../components/date";
import { GetStaticProps } from "next";
import { SearchBar } from "operand-js";

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
            <SearchBar
              apiKey={process.env.NEXT_PUBLIC_OPERAND_API_KEY}
              setId={process.env.NEXT_PUBLIC_OPERAND_SET_ID}
              feedback
              placeholderText="Search"
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                  />
                </svg>
              </div>
            </SearchBar>
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
