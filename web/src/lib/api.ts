import { loggedIn } from "$lib/store.svelte";

export async function client<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${endpoint}`, {
    ...options,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    if (response.status === 401) loggedIn.current = false;
    // Try to extract a useful error message from the backend if it exists
    const errorBody = await response.text();
    throw new Error(errorBody || `API Error: ${response.status} - ${response.statusText}`);
  }

  // Ensure the user is logged in if the request was successful
  if (!loggedIn.current) loggedIn.current = true;

  // Safely handle empty responses (like 204 No Content)
  const text = await response.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

export const api = {
  login: (secret: string) =>
    client<void>("/api/login", {
      method: "POST",
      body: JSON.stringify({ secret }),
    }),
  logout: async () => {
    await client<void>("/api/logout", { method: "POST" });
    loggedIn.current = false;
  },
  envs: () => client<string[]>("/api/envs"),
  config: (env: string) => client<any>(`/config?env=${env}`),

  events(env: string): EventSource {
    const source = new EventSource(`/api/events?env=${env}`, { withCredentials: true });
    source.onerror = (err) => {
      console.error("SSE Error", err);
    };
    return source;
  },
};
