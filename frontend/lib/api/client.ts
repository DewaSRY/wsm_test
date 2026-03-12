import axios from "axios";
import {
  AUTH_COOKIE_NAME,
  REFRESH_COOKIE_NAME,
  authResponseSchema,
  refreshRequestSchema,
} from "@/lib/auth";

function getCookieValue(name: string): string | null {
  if (typeof document === "undefined") return null;

  const target = document.cookie
    .split("; ")
    .find((cookiePart) => cookiePart.startsWith(`${name}=`));

  if (!target) return null;

  return decodeURIComponent(target.split("=").slice(1).join("="));
}

export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080",
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

const refreshClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080",
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

type ApiEnvelope<T> = {
  message: string;
  data?: T;
  error?: string;
};

function setCookieValue(name: string, value: string, maxAgeSeconds: number) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

function clearCookieValue(name: string) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=; Path=/; Max-Age=0; SameSite=Lax`;
}

function clearAuthCookies() {
  clearCookieValue(AUTH_COOKIE_NAME);
  clearCookieValue(REFRESH_COOKIE_NAME);
}

function parseEnvelope<T>(payload: unknown): T {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as ApiEnvelope<T>).data as T;
  }
  return payload as T;
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getCookieValue(REFRESH_COOKIE_NAME);
  if (!refreshToken) {
    clearAuthCookies();
    return null;
  }

  try {
    const refreshPayload = refreshRequestSchema.parse({
      refresh_token: refreshToken,
    });
    const response = await refreshClient.post("/api/auth/refresh", refreshPayload);

    const auth = authResponseSchema.parse(parseEnvelope<unknown>(response.data));
    if (!auth?.access_token || !auth?.refresh_token) return null;

    const accessTtl = auth.expires_in > 0 ? auth.expires_in : 60 * 60 * 8;
    setCookieValue(AUTH_COOKIE_NAME, auth.access_token, accessTtl);
    setCookieValue(REFRESH_COOKIE_NAME, auth.refresh_token, 60 * 60 * 24 * 30);
    return auth.access_token;
  } catch {
    clearAuthCookies();
    return null;
  }
}

apiClient.interceptors.request.use((config) => {
  const requestUrl = config.url ?? "";
  const isAuthRefreshRequest = requestUrl.includes("/api/auth/refresh");

  if (isAuthRefreshRequest) {
    return config;
  }

  const token = getCookieValue(AUTH_COOKIE_NAME);

  if (!token) {
    return config;
  }

  const headers = config.headers ?? {};

  if (typeof (headers as { set?: unknown }).set === "function") {
    (headers as { set: (key: string, value: string) => void }).set("Authorization", `Bearer ${token}`);
  } else {
    (headers as Record<string, string>).Authorization = `Bearer ${token}`;
  }

  config.headers = headers;
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error?.config as (typeof error.config & { _retry?: boolean }) | undefined;
    const status = error?.response?.status as number | undefined;

    if (status === 401 && originalRequest && !originalRequest._retry) {
      originalRequest._retry = true;
      const newAccessToken = await refreshAccessToken();

      if (newAccessToken) {
        const headers = originalRequest.headers ?? {};
        if (typeof (headers as { set?: unknown }).set === "function") {
          (headers as { set: (key: string, value: string) => void }).set("Authorization", `Bearer ${newAccessToken}`);
        } else {
          (headers as Record<string, string>).Authorization = `Bearer ${newAccessToken}`;
        }
        originalRequest.headers = headers;
        return apiClient(originalRequest);
      }

      if (typeof window !== "undefined" && window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }

    const message = error?.response?.data?.error ?? error?.response?.data?.message ?? error.message ?? "Unexpected API error";
    return Promise.reject(new Error(message));
  },
);

export function unwrapApiData<T>(payload: unknown): T {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as { data: T }).data;
  }
  return payload as T;
}