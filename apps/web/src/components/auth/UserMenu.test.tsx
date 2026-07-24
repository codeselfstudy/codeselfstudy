import { describe, expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import UserMenu from "@/components/auth/UserMenu";

const account = {
  email: "ada@example.com",
  username: "adalovelace",
  avatar: "https://example.com/ada.png",
};

describe("UserMenu", () => {
  test("the trigger shows the username and names itself for assistive tech", () => {
    render(<UserMenu account={account} onSignOut={() => {}} />);

    const trigger = screen.getByRole("button", {
      name: "Account menu for adalovelace",
    });
    expect(trigger).toBeInTheDocument();
    expect(screen.getByText("adalovelace")).toBeInTheDocument();
    // The username is shown, never the email address.
    expect(screen.queryByText("ada@example.com")).not.toBeInTheDocument();
  });

  test("the menu is closed until the trigger is clicked", async () => {
    const user = userEvent.setup();
    render(<UserMenu account={account} onSignOut={() => {}} />);

    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /account menu/i }));

    expect(await screen.findByRole("menu")).toBeInTheDocument();
    expect(
      screen.getByRole("menuitem", { name: "Settings" })
    ).toBeInTheDocument();
    expect(
      screen.getByRole("menuitem", { name: "Sign Out" })
    ).toBeInTheDocument();
  });

  test("opens on keyboard activation of the trigger", async () => {
    const user = userEvent.setup();
    render(<UserMenu account={account} onSignOut={() => {}} />);

    screen.getByRole("button", { name: /account menu/i }).focus();
    await user.keyboard("{Enter}");

    expect(await screen.findByRole("menu")).toBeInTheDocument();
  });

  test("Settings links to /settings/ with the trailing slash", async () => {
    const user = userEvent.setup();
    render(<UserMenu account={account} onSignOut={() => {}} />);

    await user.click(screen.getByRole("button", { name: /account menu/i }));

    expect(
      await screen.findByRole("menuitem", { name: "Settings" })
    ).toHaveAttribute("href", "/settings/");
  });

  test("Sign Out invokes the handler", async () => {
    const onSignOut = vi.fn();
    const user = userEvent.setup();
    render(<UserMenu account={account} onSignOut={onSignOut} />);

    await user.click(screen.getByRole("button", { name: /account menu/i }));
    await user.click(await screen.findByRole("menuitem", { name: "Sign Out" }));

    expect(onSignOut).toHaveBeenCalledTimes(1);
  });

  test("closes on Escape", async () => {
    const user = userEvent.setup();
    render(<UserMenu account={account} onSignOut={() => {}} />);

    await user.click(screen.getByRole("button", { name: /account menu/i }));
    await screen.findByRole("menu");
    await user.keyboard("{Escape}");

    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });

  test("falls back to the email label when there is no username", () => {
    render(
      <UserMenu
        account={{ email: "ada@example.com", username: "", avatar: "" }}
        onSignOut={() => {}}
      />
    );

    expect(
      screen.getByRole("button", { name: "Account menu for ada@example.com" })
    ).toBeInTheDocument();
    expect(screen.getByText("ada@example.com")).toBeInTheDocument();
  });

  test("renders the avatar when the account has one", () => {
    const { container } = render(
      <UserMenu account={account} onSignOut={() => {}} />
    );

    // Decorative (alt=""), so assert on the element rather than the a11y tree.
    expect(container.querySelector("img")).toHaveAttribute(
      "src",
      "https://example.com/ada.png"
    );
  });

  test("still names the trigger when the account is entirely empty", () => {
    render(
      <UserMenu
        account={{ email: "", username: "", avatar: "" }}
        onSignOut={() => {}}
      />
    );

    expect(
      screen.getByRole("button", { name: "Account menu" })
    ).toBeInTheDocument();
  });
});
