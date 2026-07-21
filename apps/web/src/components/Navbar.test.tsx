import { describe, expect, test } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import Navbar from "@/components/Navbar";

describe("Navbar", () => {
  test("renders the logo and desktop links with trailing-slash hrefs", () => {
    render(<Navbar />);

    expect(
      screen.getByRole("link", { name: "Code Self Study" })
    ).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "Home" })).toHaveAttribute(
      "href",
      "/"
    );
    expect(screen.getByRole("link", { name: "About" })).toHaveAttribute(
      "href",
      "/about/"
    );
    expect(screen.getByRole("link", { name: "Events" })).toHaveAttribute(
      "href",
      "/events/"
    );
  });

  test("opens the mobile drawer when the menu button is clicked", async () => {
    const user = userEvent.setup();
    render(<Navbar />);

    // The drawer content is not mounted until it is opened.
    expect(screen.queryByText("Menu")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Open main menu" }));

    expect(await screen.findByText("Menu")).toBeInTheDocument();
  });
});
