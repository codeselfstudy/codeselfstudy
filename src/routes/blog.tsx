import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/blog")({ component: Blog });

function Blog() {
  return (
    <>
      <h1>Blog</h1>
      <p>
        Lorem ipsum dolor sit, amet consectetur adipisicing elit. Vitae error
        quia autem enim accusamus nesciunt temporibus nisi quae sunt, cupiditate
        velit fugit harum assumenda sit qui eligendi. Illum, fuga nesciunt.
      </p>
    </>
  );
}
