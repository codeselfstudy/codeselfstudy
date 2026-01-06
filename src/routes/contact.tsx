import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/contact")({ component: Contact });

function Contact() {
  return (
    <>
      <h1>Contact</h1>
      <p>
        Lorem ipsum dolor sit, amet consectetur adipisicing elit. Vitae error
        quia autem enim accusamus nesciunt temporibus nisi quae sunt, cupiditate
        velit fugit harum assumenda sit qui eligendi. Illum, fuga nesciunt.
      </p>
    </>
  );
}
