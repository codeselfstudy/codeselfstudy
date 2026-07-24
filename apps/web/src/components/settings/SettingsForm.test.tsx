import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import SettingsForm from "@/components/settings/SettingsForm";

// SettingsForm drives every state off the cookie-gated /api/settings, so each
// test stubs fetch: the first call is the on-mount load, later calls are the
// PATCH (save) or the delete-request POST. Location is reset per test because
// welcome mode reads and rewrites window.location.

type Body = Record<string, unknown> | null;

function res(status: number, data: Body) {
  return { ok: status >= 200 && status < 300, status, json: async () => data };
}

function settings(over: Body = {}) {
  return res(200, {
    username: "ada",
    email: "ada@example.com",
    timezone: "America/Los_Angeles",
    deletion_requested_at: null,
    ...over,
  });
}

const unauthorized = res(401, null);

function mockFetch(...responses: Array<ReturnType<typeof res>>) {
  const fn = vi.fn();
  for (const r of responses) fn.mockResolvedValueOnce(r);
  vi.stubGlobal("fetch", fn);
  return fn;
}

beforeEach(() => {
  window.history.replaceState(null, "", "/settings/");
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("SettingsForm", () => {
  test("loading: shows a placeholder and no form until /api/settings resolves", () => {
    vi.stubGlobal("fetch", vi.fn().mockReturnValue(new Promise(() => {})));
    const { container } = render(<SettingsForm />);

    expect(container.querySelector(".animate-pulse")).toBeInTheDocument();
    expect(screen.queryByLabelText("Username")).not.toBeInTheDocument();
  });

  test("signed out (401): shows the sign-in panel linking to /auth/login with returnTo", async () => {
    mockFetch(unauthorized);
    render(<SettingsForm />);

    expect(
      await screen.findByText("Sign in to manage your account.")
    ).toBeInTheDocument();
    const link = screen.getByRole("link", { name: "Sign In" });
    expect(link).toHaveAttribute("href", "/auth/login?returnTo=%2Fsettings%2F");
    expect(screen.queryByLabelText("Username")).not.toBeInTheDocument();
  });

  test("load failure (500): shows an error, not the sign-in panel", async () => {
    mockFetch(res(500, null));
    render(<SettingsForm />);

    expect(
      await screen.findByText(/couldn’t load your settings/i)
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Sign in to manage your account.")
    ).not.toBeInTheDocument();
  });

  test("ready: renders the loaded username, read-only email, and timezone", async () => {
    mockFetch(settings());
    render(<SettingsForm />);

    expect(await screen.findByLabelText("Username")).toHaveValue("ada");
    const email = screen.getByLabelText("Email");
    expect(email).toHaveValue("ada@example.com");
    expect(email).toHaveAttribute("readonly");
    expect(screen.getByLabelText("Time zone")).toHaveValue(
      "America/Los_Angeles"
    );
  });

  test("requests /api/settings with same-origin credentials so the cookie is sent", async () => {
    const fetchMock = mockFetch(settings());
    render(<SettingsForm />);

    await screen.findByLabelText("Username");
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/settings",
      expect.objectContaining({ credentials: "same-origin" })
    );
  });

  test("empty timezone from the API falls back to the browser zone", async () => {
    mockFetch(settings({ timezone: "" }));
    render(<SettingsForm />);

    const select = (await screen.findByLabelText(
      "Time zone"
    )) as HTMLSelectElement;
    expect(select.value).not.toBe("");
  });

  test("saving PATCHes the new username and shows Saved", async () => {
    const fetchMock = mockFetch(
      settings(),
      res(200, {
        username: "adalovelace",
        email: "ada@example.com",
        timezone: "America/Los_Angeles",
        deletion_requested_at: null,
      })
    );
    const user = userEvent.setup();
    render(<SettingsForm />);

    const input = await screen.findByLabelText("Username");
    await user.clear(input);
    await user.type(input, "adalovelace");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByRole("status")).toHaveTextContent("Saved");
    const patch = fetchMock.mock.calls.find(
      (c) => (c[1] as RequestInit | undefined)?.method === "PATCH"
    );
    expect(patch).toBeTruthy();
    const opts = patch![1] as RequestInit;
    expect(opts.credentials).toBe("same-origin");
    expect(JSON.parse(opts.body as string)).toEqual({
      username: "adalovelace",
      timezone: "America/Los_Angeles",
    });
  });

  test("409 shows 'That username is taken'", async () => {
    mockFetch(settings(), res(409, { error: "username_taken" }));
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.clear(await screen.findByLabelText("Username"));
    await user.type(screen.getByLabelText("Username"), "taken");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(
      await screen.findByText("That username is taken")
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Username")).toHaveAttribute(
      "aria-invalid",
      "true"
    );
  });

  test("429 shows the retry window from retry_after_days", async () => {
    mockFetch(
      settings(),
      res(429, { error: "rate_limited", retry_after_days: 12 })
    );
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.clear(await screen.findByLabelText("Username"));
    await user.type(screen.getByLabelText("Username"), "toosoon");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(
      await screen.findByText("You can change your username again in 12 days")
    ).toBeInTheDocument();
  });

  test("400 invalid username shows the character-set hint", async () => {
    mockFetch(settings(), res(400, { error: "username_invalid" }));
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.clear(await screen.findByLabelText("Username"));
    await user.type(screen.getByLabelText("Username"), "bad name");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(
      await screen.findByText(/letters, numbers, hyphens and underscores/i)
    ).toBeInTheDocument();
  });

  test("a pending deletion from the API shows the pending state, not the button", async () => {
    mockFetch(settings({ deletion_requested_at: "2026-07-01T00:00:00Z" }));
    render(<SettingsForm />);

    expect(
      await screen.findByText(/you’ve requested account deletion/i)
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Request account deletion" })
    ).not.toBeInTheDocument();
  });

  test("requesting deletion confirms, POSTs, then shows the pending state", async () => {
    const fetchMock = mockFetch(
      settings(),
      res(202, { deletion_requested_at: "2026-07-24T00:00:00Z" })
    );
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(
      await screen.findByRole("button", { name: "Request account deletion" })
    );
    // Confirmation copy is explicit that an admin actions it manually.
    expect(
      screen.getByText(/deletes the account manually/i)
    ).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: "Yes, request deletion" })
    );

    expect(
      await screen.findByText(/you’ve requested account deletion/i)
    ).toBeInTheDocument();
    const post = fetchMock.mock.calls.find(
      (c) => (c[1] as RequestInit | undefined)?.method === "POST"
    );
    expect(post?.[0]).toBe("/api/settings/delete-request");
  });

  test("cancelling the deletion confirm returns to the request button", async () => {
    mockFetch(settings());
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(
      await screen.findByRole("button", { name: "Request account deletion" })
    );
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(
      screen.getByRole("button", { name: "Request account deletion" })
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Yes, request deletion" })
    ).not.toBeInTheDocument();
  });

  test("welcome mode focuses and selects the username, then strips ?welcome=1", async () => {
    window.history.replaceState(null, "", "/settings/?welcome=1");
    mockFetch(settings({ username: "clever-otter-42" }));
    render(<SettingsForm />);

    expect(
      await screen.findByRole("heading", { name: "Pick a username" })
    ).toBeInTheDocument();
    const input = (await screen.findByLabelText(
      "Username"
    )) as HTMLInputElement;

    await waitFor(() => expect(document.activeElement).toBe(input));
    expect(input.selectionStart).toBe(0);
    expect(input.selectionEnd).toBe(input.value.length);
    expect(window.location.search).toBe("");
  });

  test("timezone select still offers the current zone when the runtime lacks supportedValuesOf", async () => {
    const original = Intl.supportedValuesOf;
    // @ts-expect-error — simulate an older runtime without the API.
    Intl.supportedValuesOf = undefined;
    try {
      mockFetch(settings({ timezone: "America/Los_Angeles" }));
      render(<SettingsForm />);

      const select = (await screen.findByLabelText(
        "Time zone"
      )) as HTMLSelectElement;
      expect(select.value).toBe("America/Los_Angeles");
      expect(
        screen.getByRole("option", { name: "America/Los_Angeles" })
      ).toBeInTheDocument();
    } finally {
      Intl.supportedValuesOf = original;
    }
  });

  test("400 invalid timezone shows the timezone error", async () => {
    mockFetch(settings(), res(400, { error: "timezone_invalid" }));
    const user = userEvent.setup();
    render(<SettingsForm />);

    await screen.findByLabelText("Time zone");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(
      await screen.findByText("That time zone isn’t valid")
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Time zone")).toHaveAttribute(
      "aria-invalid",
      "true"
    );
  });

  test("429 without retry_after_days falls back to the default window", async () => {
    mockFetch(settings(), res(429, { error: "rate_limited" }));
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(await screen.findByRole("button", { name: "Save" }));

    expect(
      await screen.findByText("You can change your username again in 30 days")
    ).toBeInTheDocument();
  });

  test("an unmapped save failure shows a generic error", async () => {
    mockFetch(settings(), res(500, null));
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(await screen.findByRole("button", { name: "Save" }));

    expect(
      await screen.findByText(/couldn’t save your settings/i)
    ).toBeInTheDocument();
  });

  test("a network error while saving shows a generic error", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(settings())
      .mockRejectedValueOnce(new Error("offline"));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(await screen.findByRole("button", { name: "Save" }));

    expect(
      await screen.findByText(/couldn’t save your settings/i)
    ).toBeInTheDocument();
  });

  test("a failed deletion request surfaces an error and keeps the confirm open", async () => {
    mockFetch(settings(), res(500, null));
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(
      await screen.findByRole("button", { name: "Request account deletion" })
    );
    await user.click(
      screen.getByRole("button", { name: "Yes, request deletion" })
    );

    expect(
      await screen.findByText(/couldn’t submit your request/i)
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Yes, request deletion" })
    ).toBeInTheDocument();
  });

  test("a network error while requesting deletion surfaces an error", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(settings())
      .mockRejectedValueOnce(new Error("offline"));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    render(<SettingsForm />);

    await user.click(
      await screen.findByRole("button", { name: "Request account deletion" })
    );
    await user.click(
      screen.getByRole("button", { name: "Yes, request deletion" })
    );

    expect(
      await screen.findByText(/couldn’t submit your request/i)
    ).toBeInTheDocument();
  });
});
