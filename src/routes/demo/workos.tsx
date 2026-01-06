import { createFileRoute } from "@tanstack/react-router";
import { useAuth } from "@workos-inc/authkit-react";

export const Route = createFileRoute("/demo/workos")({
  ssr: false,
  component: App,
});

function App() {
  const { user, isLoading, signIn, signOut } = useAuth();

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 p-4">
        <div className="w-full max-w-md rounded-2xl border border-gray-700/50 bg-gray-800/50 p-8 shadow-2xl backdrop-blur-sm">
          <p className="text-center text-gray-400">Loading...</p>
        </div>
      </div>
    );
  }

  if (user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 p-4">
        <div className="w-full max-w-md rounded-2xl border border-gray-700/50 bg-gray-800/50 p-8 shadow-2xl backdrop-blur-sm">
          <h1 className="mb-6 text-center text-2xl font-bold text-white">
            User Profile
          </h1>

          <div className="space-y-6">
            {/* Profile Picture */}
            {user.profilePictureUrl && (
              <div className="flex justify-center">
                <img
                  src={user.profilePictureUrl}
                  alt={`Avatar of ${user.firstName} ${user.lastName}`}
                  className="h-24 w-24 rounded-full border-4 border-gray-700 shadow-lg"
                />
              </div>
            )}

            {/* User Information */}
            <div className="space-y-4">
              <div className="rounded-lg border border-gray-600/30 bg-gray-700/30 p-4">
                <label className="mb-1 block text-sm font-medium text-gray-400">
                  First Name
                </label>
                <p className="text-lg text-white">{user.firstName || "N/A"}</p>
              </div>

              <div className="rounded-lg border border-gray-600/30 bg-gray-700/30 p-4">
                <label className="mb-1 block text-sm font-medium text-gray-400">
                  Last Name
                </label>
                <p className="text-lg text-white">{user.lastName || "N/A"}</p>
              </div>

              <div className="rounded-lg border border-gray-600/30 bg-gray-700/30 p-4">
                <label className="mb-1 block text-sm font-medium text-gray-400">
                  Email
                </label>
                <p className="text-lg break-all text-white">
                  {user.email || "N/A"}
                </p>
              </div>

              <div className="rounded-lg border border-gray-600/30 bg-gray-700/30 p-4">
                <label className="mb-1 block text-sm font-medium text-gray-400">
                  User ID
                </label>
                <p className="font-mono text-sm break-all text-gray-300">
                  {user.id || "N/A"}
                </p>
              </div>
            </div>

            {/* Sign Out Button */}
            <button
              onClick={() => signOut()}
              className="w-full rounded-lg bg-blue-600 px-6 py-3 font-medium text-white shadow-lg transition-colors hover:bg-blue-700 hover:shadow-xl"
            >
              Sign Out
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 p-4">
      <div className="w-full max-w-md rounded-2xl border border-gray-700/50 bg-gray-800/50 p-8 shadow-2xl backdrop-blur-sm">
        <h1 className="mb-6 text-center text-2xl font-bold text-white">
          WorkOS Authentication
        </h1>
        <p className="mb-6 text-center text-gray-400">
          Sign in to view your profile information
        </p>
        <button
          onClick={() => signIn()}
          disabled={isLoading}
          className="w-full rounded-lg bg-blue-600 px-6 py-3 font-medium text-white shadow-lg transition-colors hover:bg-blue-700 hover:shadow-xl disabled:cursor-not-allowed disabled:opacity-50"
        >
          Sign In with AuthKit
        </button>
      </div>
    </div>
  );
}
