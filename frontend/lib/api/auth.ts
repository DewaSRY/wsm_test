import axios from "axios";
import {
  AUTH_COOKIE_MAX_AGE_SECONDS,
  AUTH_COOKIE_NAME,
  REFRESH_COOKIE_MAX_AGE_SECONDS,
  REFRESH_COOKIE_NAME,
  authResponseSchema,
  loginRequestSchema,
  registerRequestSchema,
  type AuthResponse,
  type LoginRequest,
  type RegisterRequest,
} from "@/lib/auth";

type ApiEnvelope<T> = {
  message: string;
  data?: T;
  error?: string;
};

const authHttp = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080",
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

function setCookie(name: string, value: string, maxAgeSeconds: number) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`;
}

function clearCookie(name: string) {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=; Path=/; Max-Age=0; SameSite=Lax`;
}

function getResponseData<T>(payload: unknown): T {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as ApiEnvelope<T>).data as T;
  }
  return payload as T;
}

export function setAuthSession(auth: AuthResponse) {
  const accessTtl =
    auth.expires_in > 0 ? auth.expires_in : AUTH_COOKIE_MAX_AGE_SECONDS;
  setCookie(AUTH_COOKIE_NAME, auth.access_token, accessTtl);
  setCookie(
    REFRESH_COOKIE_NAME,
    auth.refresh_token,
    REFRESH_COOKIE_MAX_AGE_SECONDS,
  );
}

export function clearAuthSession() {
  clearCookie(AUTH_COOKIE_NAME);
  clearCookie(REFRESH_COOKIE_NAME);
}

export async function registerUser(
  input: RegisterRequest,
): Promise<AuthResponse> {
  const parsedInput = registerRequestSchema.parse(input);
  await authHttp.post("/api/auth/register", parsedInput);
  return loginUser({ email: parsedInput.email, password: parsedInput.password });
}

export async function loginUser(input: LoginRequest): Promise<AuthResponse> {
  const parsedInput = loginRequestSchema.parse(input);
  const response = await authHttp.post("/api/auth/login", parsedInput);
  return authResponseSchema.parse(getResponseData<unknown>(response.data));
}
