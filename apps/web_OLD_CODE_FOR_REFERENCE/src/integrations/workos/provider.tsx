import { AuthKitProvider } from "@workos-inc/authkit-react";
import { useNavigate } from "@tanstack/react-router";

import { env } from "@/env";

const VITE_WORKOS_CLIENT_ID = env.VITE_WORKOS_CLIENT_ID;
const VITE_WORKOS_API_HOSTNAME = env.VITE_WORKOS_API_HOSTNAME;

export default function AppWorkOSProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const navigate = useNavigate();

  return (
    <AuthKitProvider
      clientId={VITE_WORKOS_CLIENT_ID}
      apiHostname={VITE_WORKOS_API_HOSTNAME}
      onRedirectCallback={({ state }) => {
        if (state?.returnTo) {
          navigate(state.returnTo);
        }
      }}
    >
      {children}
    </AuthKitProvider>
  );
}
