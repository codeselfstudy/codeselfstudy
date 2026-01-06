import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/about")({ component: About });

function About() {
  return (
    <>
      <h1>About</h1>
      <p>
        Lorem ipsum dolor sit, amet consectetur adipisicing elit. Vitae error
        quia autem enim accusamus nesciunt temporibus nisi quae sunt, cupiditate
        velit fugit harum assumenda sit qui eligendi. Illum, fuga nesciunt.
      </p>
    </>
  );
}
