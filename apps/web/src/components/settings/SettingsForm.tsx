import { useEffect, useMemo, useRef, useState } from "react";
import type { FormEvent } from "react";

import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// SettingsForm is the single React island behind the static /settings/ shell.
// The site is prerendered, so there is no per-user server render: this component
// asks the cookie-gated /api/settings who the caller is and renders its own
// signed-out state, mirroring SignInButton. It reads the DB-backed settings
// (username is the display name, email is read-only from WorkOS, timezone), lets
// the user save changes (PATCH, with the store's 400/409/429 mapped to inline
// copy), and files a manual account-deletion request. In welcome mode
// (?welcome=1, set by the server right after sign-up) it focuses and selects the
// auto-generated username so a single keystroke replaces it.

type Settings = {
  username: string;
  email: string;
  timezone: string;
  deletion_requested_at: string | null;
};

type Status = "loading" | "signedout" | "error" | "ready";
type ErrorField = "username" | "timezone" | null;
type SaveError = { field: ErrorField; text: string };

// browserTimezone is the default for an account that has never set one — the
// visitor's own zone, resolved by the runtime.
function browserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

// timezoneOptions is the full IANA list the runtime knows, with `current`
// guaranteed present so a controlled <select> always has its value as an option
// (a runtime without supportedValuesOf still offers at least the current zone).
function timezoneOptions(current: string): string[] {
  const list =
    typeof Intl.supportedValuesOf === "function"
      ? Intl.supportedValuesOf("timeZone")
      : [];
  if (current && !list.includes(current)) return [current, ...list];
  return list;
}

// loginHref points at the server's hosted sign-in, carrying the current in-app
// location as returnTo so the flow lands the user back here. The server
// re-validates returnTo to a same-origin path (session.safeReturnTo), so a
// crafted value can't turn this into an open redirect.
function loginHref(): string {
  const returnTo =
    window.location.pathname + window.location.search + window.location.hash;
  return `/auth/login?returnTo=${encodeURIComponent(returnTo)}`;
}

// saveError maps a failed PATCH /api/settings to the field it concerns and the
// copy shown inline. The status-code contract with the API lives here, in one
// place: 409 taken, 429 within the rename cooldown (with retry_after_days), 400
// reserved/invalid username or invalid timezone.
function saveError(
  status: number,
  body: { error?: string; retry_after_days?: number } | null
): SaveError {
  if (status === 409) {
    return { field: "username", text: "That username is taken" };
  }
  if (status === 429) {
    const n = body?.retry_after_days ?? 30;
    return {
      field: "username",
      text: `You can change your username again in ${n} ${n === 1 ? "day" : "days"}`,
    };
  }
  if (status === 400) {
    switch (body?.error) {
      case "username_reserved":
        return { field: "username", text: "That name is reserved" };
      case "username_invalid":
        return {
          field: "username",
          text: "Usernames can use letters, numbers, hyphens and underscores",
        };
      case "timezone_invalid":
        return { field: "timezone", text: "That time zone isn’t valid" };
    }
  }
  return {
    field: null,
    text: "Couldn’t save your settings. Please try again.",
  };
}

const inputClass =
  "border-input bg-background focus-visible:ring-ring/50 aria-invalid:border-destructive w-full rounded-md border px-3 py-2 text-sm focus-visible:ring-[3px] focus-visible:outline-none";

export default function SettingsForm() {
  const [status, setStatus] = useState<Status>("loading");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [timezone, setTimezone] = useState("");
  const [deletionRequestedAt, setDeletionRequestedAt] = useState<string | null>(
    null
  );

  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<SaveError | null>(null);

  const [confirmingDelete, setConfirmingDelete] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const usernameRef = useRef<HTMLInputElement>(null);
  // Captured once, before the welcome effect strips it from the URL.
  const [welcome] = useState(
    () => new URLSearchParams(window.location.search).get("welcome") === "1"
  );

  // Load the current settings once. 401 -> signed out; any other non-2xx or a
  // network error -> a transient error state (a 500 must not masquerade as
  // "please sign in").
  useEffect(() => {
    let cancelled = false;
    fetch("/api/settings", {
      credentials: "same-origin",
      headers: { Accept: "application/json" },
    })
      .then(async (res) => {
        if (cancelled) return;
        if (res.status === 401) {
          setStatus("signedout");
          return;
        }
        if (!res.ok) {
          setStatus("error");
          return;
        }
        const data = (await res.json()) as Partial<Settings>;
        if (cancelled) return;
        setEmail(data.email ?? "");
        setUsername(data.username ?? "");
        setTimezone(data.timezone || browserTimezone());
        setDeletionRequestedAt(data.deletion_requested_at ?? null);
        setStatus("ready");
      })
      .catch(() => {
        if (!cancelled) setStatus("error");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Welcome mode: once the data is in, focus the username and select its text so
  // a single keystroke replaces the auto-generated name. Runs after load (keyed
  // on status), not on mount. Strip ?welcome=1 so a reload doesn't re-trigger.
  useEffect(() => {
    if (status !== "ready" || !welcome) return;
    const el = usernameRef.current;
    if (el) {
      el.focus();
      el.select();
    }
    const url = new URL(window.location.href);
    url.searchParams.delete("welcome");
    window.history.replaceState(null, "", url.pathname + url.search + url.hash);
  }, [status, welcome]);

  const zones = useMemo(() => timezoneOptions(timezone), [timezone]);

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaved(false);
    setError(null);
    try {
      const res = await fetch("/api/settings", {
        method: "PATCH",
        credentials: "same-origin",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, timezone }),
      });
      if (res.ok) {
        const data = (await res.json()) as Partial<Settings>;
        setUsername(data.username ?? username);
        setTimezone(data.timezone || timezone);
        setDeletionRequestedAt(data.deletion_requested_at ?? null);
        setSaved(true);
      } else {
        const body = (await res.json().catch(() => null)) as {
          error?: string;
          retry_after_days?: number;
        } | null;
        setError(saveError(res.status, body));
      }
    } catch {
      setError({
        field: null,
        text: "Couldn’t save your settings. Please try again.",
      });
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    setDeleting(true);
    setError(null);
    try {
      const res = await fetch("/api/settings/delete-request", {
        method: "POST",
        credentials: "same-origin",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({}),
      });
      if (res.ok) {
        const data = (await res.json().catch(() => null)) as {
          deletion_requested_at?: string | null;
        } | null;
        setDeletionRequestedAt(data?.deletion_requested_at ?? null);
        setConfirmingDelete(false);
      } else {
        setError({
          field: null,
          text: "Couldn’t submit your request. Please try again.",
        });
      }
    } catch {
      setError({
        field: null,
        text: "Couldn’t submit your request. Please try again.",
      });
    } finally {
      setDeleting(false);
    }
  }

  // Any edit clears the last save's result so stale "Saved"/error copy never
  // lingers over changed fields.
  function onEdit() {
    setSaved(false);
    setError(null);
  }

  if (status === "loading") {
    return (
      <div
        className="bg-muted h-64 animate-pulse rounded-md"
        aria-hidden="true"
      />
    );
  }

  if (status === "signedout") {
    return (
      <div className="mx-auto max-w-md text-center">
        <h1 className="text-2xl font-bold tracking-tight">Account settings</h1>
        <p className="text-muted-foreground mt-2 text-sm">
          Sign in to manage your account.
        </p>
        <a href={loginHref()} className={cn(buttonVariants(), "mt-6")}>
          Sign In
        </a>
      </div>
    );
  }

  if (status === "error") {
    return (
      <div className="mx-auto max-w-md text-center">
        <h1 className="text-2xl font-bold tracking-tight">Account settings</h1>
        <p className="text-muted-foreground mt-2 text-sm" role="alert">
          We couldn’t load your settings. Please reload the page.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-10">
      <header>
        <h1 className="text-2xl font-bold tracking-tight">
          {welcome ? "Pick a username" : "Account settings"}
        </h1>
        <p className="text-muted-foreground mt-2 text-sm">
          {welcome
            ? "We gave you a starter username. Make it yours — change it now, then again later if you like."
            : "Manage how you appear on Code Self Study."}
        </p>
      </header>

      <form onSubmit={handleSave} className="space-y-6" noValidate>
        <div className="space-y-1.5">
          <label htmlFor="username" className="block text-sm font-medium">
            Username
          </label>
          <input
            id="username"
            ref={usernameRef}
            type="text"
            value={username}
            onChange={(e) => {
              setUsername(e.target.value);
              onEdit();
            }}
            autoComplete="off"
            autoCapitalize="none"
            spellCheck={false}
            aria-invalid={error?.field === "username" || undefined}
            aria-describedby={
              error?.field === "username" ? "username-error" : undefined
            }
            className={inputClass}
          />
          <p className="text-muted-foreground text-xs">
            This is also your display name.
          </p>
          {error?.field === "username" && (
            <p
              id="username-error"
              role="alert"
              className="text-destructive text-sm"
            >
              {error.text}
            </p>
          )}
        </div>

        <div className="space-y-1.5">
          <label htmlFor="email" className="block text-sm font-medium">
            Email
          </label>
          <input
            id="email"
            type="email"
            value={email}
            readOnly
            className={cn(inputClass, "bg-muted text-muted-foreground")}
          />
          <p className="text-muted-foreground text-xs">
            Managed by your sign-in provider.
          </p>
        </div>

        <div className="space-y-1.5">
          <label htmlFor="timezone" className="block text-sm font-medium">
            Time zone
          </label>
          <select
            id="timezone"
            value={timezone}
            onChange={(e) => {
              setTimezone(e.target.value);
              onEdit();
            }}
            aria-invalid={error?.field === "timezone" || undefined}
            className={inputClass}
          >
            {zones.map((z) => (
              <option key={z} value={z}>
                {z}
              </option>
            ))}
          </select>
          {error?.field === "timezone" && (
            <p role="alert" className="text-destructive text-sm">
              {error.text}
            </p>
          )}
        </div>

        {error?.field === null && (
          <p role="alert" className="text-destructive text-sm">
            {error.text}
          </p>
        )}

        <div className="flex items-center gap-3">
          <button
            type="submit"
            disabled={saving}
            className={cn(buttonVariants())}
          >
            {saving ? "Saving…" : "Save"}
          </button>
          {saved && (
            <span role="status" className="text-muted-foreground text-sm">
              Saved
            </span>
          )}
        </div>
      </form>

      <section className="border-destructive/30 rounded-md border p-4">
        <h2 className="text-sm font-semibold">Delete account</h2>
        {deletionRequestedAt ? (
          <p className="text-muted-foreground mt-2 text-sm" role="status">
            You’ve requested account deletion. An admin actions this manually —
            your account has not been deleted yet.
          </p>
        ) : confirmingDelete ? (
          <div className="mt-2 space-y-3">
            <p className="text-muted-foreground text-sm">
              Are you sure? This sends a request to an admin, who deletes the
              account manually. Nothing is deleted immediately, and you can keep
              using your account until then.
            </p>
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleting}
                className={cn(buttonVariants({ variant: "destructive" }))}
              >
                {deleting ? "Requesting…" : "Yes, request deletion"}
              </button>
              <button
                type="button"
                onClick={() => setConfirmingDelete(false)}
                disabled={deleting}
                className={cn(buttonVariants({ variant: "outline" }))}
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <>
            <p className="text-muted-foreground mt-2 text-sm">
              Request that an admin delete your account. This is manual and not
              immediate.
            </p>
            <button
              type="button"
              onClick={() => setConfirmingDelete(true)}
              className={cn(buttonVariants({ variant: "destructive" }), "mt-3")}
            >
              Request account deletion
            </button>
          </>
        )}
      </section>
    </div>
  );
}
