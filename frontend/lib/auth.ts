import { z } from "zod";

export const AUTH_COOKIE_NAME = "wms_session";
export const REFRESH_COOKIE_NAME = "wms_refresh";

// Keep session cookie lifetime aligned with backend access token defaults.
export const AUTH_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 8;

// Refresh token lifetime is typically longer than access token.
export const REFRESH_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 30;

export const authUserSchema = z.object({
	id: z.number(),
	email: z.string().email(),
	name: z.string(),
});

export const authResponseSchema = z.object({
	access_token: z.string().min(1),
	refresh_token: z.string().min(1),
	expires_in: z.number().int().nonnegative(),
	token_type: z.string().min(1),
	user: authUserSchema,
});

export const loginRequestSchema = z.object({
	email: z.string().email(),
	password: z.string().min(6),
});

export const registerRequestSchema = z.object({
	email: z.string().email(),
	password: z.string().min(6),
	name: z.string().min(2),
});

export const refreshRequestSchema = z.object({
	refresh_token: z.string().min(1),
});

export type AuthUser = z.infer<typeof authUserSchema>;
export type AuthResponse = z.infer<typeof authResponseSchema>;
export type LoginRequest = z.infer<typeof loginRequestSchema>;
export type RegisterRequest = z.infer<typeof registerRequestSchema>;
export type RefreshRequest = z.infer<typeof refreshRequestSchema>;