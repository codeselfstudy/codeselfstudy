import { renderHook } from "@testing-library/react";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
  vi,
} from "vitest";
import { useUser } from "./use-user";

const { useAuthMock, useLocationMock } = vi.hoisted(() => ({
  useAuthMock: vi.fn(),
  useLocationMock: vi.fn(),
}));

vi.mock("@workos-inc/authkit-react", () => ({
  useAuth: () => useAuthMock(),
}));

vi.mock("@tanstack/react-router", () => ({
  useLocation: () => useLocationMock(),
}));

type AuthState = {
  user: { id: string; email: string } | null;
  isLoading: boolean;
  signIn: ReturnType<typeof vi.fn>;
};

function configureAuth(state: Partial<AuthState> = {}): AuthState {
  const auth: AuthState = {
    user: null,
    isLoading: false,
    signIn: vi.fn(),
    ...state,
  };
  useAuthMock.mockReturnValue(auth);
  return auth;
}

function configureLocation(pathname = "/some/path") {
  useLocationMock.mockReturnValue({ pathname });
}

describe("useUser", () => {
  let logSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    useAuthMock.mockReset();
    useLocationMock.mockReset();
    configureLocation();
    logSpy = vi.spyOn(console, "log").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("does not redirect while auth is loading", () => {
    const auth = configureAuth({ user: null, isLoading: true });

    const { result } = renderHook(() => useUser());

    expect(result.current).toBeNull();
    expect(auth.signIn).not.toHaveBeenCalled();
    // While loading, the effect's else-branch logs the still-null user.
    expect(logSpy).toHaveBeenCalledWith(null);
  });

  test("calls signIn with the current pathname when no user is present", () => {
    const auth = configureAuth({ user: null, isLoading: false });
    configureLocation("/protected/area");

    const { result } = renderHook(() => useUser());

    expect(result.current).toBeNull();
    expect(auth.signIn).toHaveBeenCalledTimes(1);
    expect(auth.signIn).toHaveBeenCalledWith({
      state: { returnTo: "/protected/area" },
    });
  });

  test("returns the authenticated user without redirecting", () => {
    const user = { id: "u_1", email: "alice@example.com" };
    const auth = configureAuth({ user, isLoading: false });

    const { result } = renderHook(() => useUser());

    expect(result.current).toEqual(user);
    expect(auth.signIn).not.toHaveBeenCalled();
    expect(logSpy).toHaveBeenCalledWith(user);
  });

  test("re-runs the effect when isLoading flips from true to false without a user", () => {
    const signIn = vi.fn();
    useAuthMock.mockReturnValue({ user: null, isLoading: true, signIn });

    const { rerender } = renderHook(() => useUser());
    expect(signIn).not.toHaveBeenCalled();

    useAuthMock.mockReturnValue({ user: null, isLoading: false, signIn });
    rerender();

    expect(signIn).toHaveBeenCalledTimes(1);
  });

  test("does not re-run signIn on a rerender with stable inputs", () => {
    const auth = configureAuth({ user: null, isLoading: false });
    const { rerender } = renderHook(() => useUser());
    expect(auth.signIn).toHaveBeenCalledTimes(1);
    rerender();
    expect(auth.signIn).toHaveBeenCalledTimes(1);
  });
});
