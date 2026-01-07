import { Link } from "@tanstack/react-router";
import { cn } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

export function HomeContent() {
  return (
    <>
      <div className="bg-circuit-board hero homepage-hero bg-cover bg-center py-20">
        <div className="container mx-auto bg-white p-4 px-4 text-center">
          <h1>Code Self Study</h1>
          <h2 className="font-light">
            Self-study computer programming in a community of like-minded people
          </h2>
          <p className="mx-auto mb-8 max-w-2xl text-lg">
            We're a friendly programming community made up of over 5,000
            programmers in Berkeley and the San Francisco Bay Area. All
            programming languages and levels of experience are welcome!
          </p>
          <a
            href="https://www.meetup.com/codeselfstudy/"
            target="_blank"
            rel="noopener noreferrer nofollow"
            className={cn(
              buttonVariants({ size: "lg" }),
              "bg-blue-600 text-lg text-white hover:bg-blue-700"
            )}
          >
            Join the Community
          </a>
        </div>
      </div>

      <section className="py-12">
        <div className="container mx-auto px-4">
          <div className="grid gap-8 md:grid-cols-3">
            <div>
              <h2>Community</h2>
              <p>
                Attend our meetups and get out of your room to meet like-minded
                programmers.
              </p>
            </div>
            <div>
              <h2>Get Help</h2>
              <p>
                The experienced members in the group help the beginners. Bring
                your programming questions, and don't hesitate to ask. We can
                also help you design study plans. The model is self-study with
                mentorship.
              </p>
            </div>
            <div>
              <h2>Networking</h2>
              <p>
                It's easier to get a job when you know people in the industry.
                Many people in the group have gotten jobs through networking or
                through tips they've acquired from other group members.
              </p>
            </div>
          </div>
          <hr className="my-12 border-gray-200" />
        </div>
      </section>

      <section className="pb-12">
        <div className="container mx-auto px-4">
          <h2>What We Do</h2>
          <p>
            There are dozens of local meetup groups that teach people how to
            code. Here are some ways that we are different from most other
            groups:
          </p>
          <ul>
            <li>
              We are building a local community of friendly, creative,
              highly-motivated programmers in San Francisco Bay Area.
            </li>
            <li>
              We're open to all programming languages, all levels of ability,
              and all types of computer programming.
            </li>
            <li>
              The format is self study with advice and mentorship from the more
              experienced members. We can help you with study plans and guidance
              &mdash; just let us know at a meetup or in the forum.
            </li>
          </ul>
          <p>
            Some of the languages our members have been working with are:
            Python, JavaScript, TypeScript, Go, Rust, WebAssembly, Java, Scheme,
            Racket, Haskell, Clojure, Common Lisp, C, C++, C#, F#, Lua, Elixir,
            Erlang, Elm, Purescript, Ruby, Scala, PHP, R, Julia, HTML/CSS,
            Octave, SQL and more. We're open to anything that is
            computer-related.
          </p>

          <h2>Coding Bootcamp Supplement</h2>
          <p>
            You can attend our meetings as a self-directed, coding bootcamp
            supplement or alternative. Experienced programmers in the group can
            offer mentorship on a project-based approach to learning.
          </p>
          <p>
            This community is self-directed &mdash; if you would like assistance
            with forming a study plan, let us know in the forum, chatroom,
            and/or at the meetups.
          </p>
        </div>
      </section>
    </>
  );
}
