"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Box, Lock, User } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { loginUser, setAuthSession } from "@/lib/api/auth";

const loginSchema = z.object({
  email: z.string().email("Valid email is required"),
  password: z.string().min(6, "Password must be at least 6 characters"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const [submitError, setSubmitError] = useState<string | null>(null);

  const form = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const onSubmit = async (values: LoginFormValues) => {
    setSubmitError(null);
    try {
      const auth = await loginUser({
        email: values.email,
        password: values.password,
      });
      setAuthSession(auth);
      router.push("/");
      router.refresh();
    } catch (error) {
      setSubmitError(error instanceof Error ? error.message : "Login failed");
    }
  };

  return (
    <main className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,#e8efff_0,#f3f6ff_44%,#f7f8fc_100%)] px-4">
      <section className="w-full max-w-md rounded-2xl border border-[#dfe4ff] bg-white p-6 shadow-[0_20px_40px_-25px_rgba(25,56,154,0.45)]">
        <div className="mb-6">
          <div className="inline-flex items-center gap-2 rounded-full bg-[#eef2ff] px-3 py-1 text-xs font-medium text-[#355dd8]">
            <Box className="h-3.5 w-3.5" />
            WMSpaceIO
          </div>
          <h1 className="mt-3 font-[var(--font-space-grotesk)] text-3xl font-semibold text-[#1d2a53]">Sign in</h1>
          <p className="mt-1 text-sm text-[#5f6b90]">Login to access outbound table and warehouse controls.</p>
        </div>

        <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
          <div>
            <label className="mb-1 block text-sm font-medium text-[#2f3b63]">Email</label>
            <div className="relative">
              <User className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8691b4]" />
              <Input
                className="h-11 border-[#d6ddff] pl-9"
                placeholder="warehouse@example.com"
                autoComplete="email"
                {...form.register("email")}
              />
            </div>
            {form.formState.errors.email ? (
              <p className="mt-1 text-xs text-[#d74747]">{form.formState.errors.email.message}</p>
            ) : null}
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-[#2f3b63]">Password</label>
            <div className="relative">
              <Lock className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8691b4]" />
              <Input
                type="password"
                className="h-11 border-[#d6ddff] pl-9"
                placeholder="Enter password"
                autoComplete="current-password"
                {...form.register("password")}
              />
            </div>
            {form.formState.errors.password ? (
              <p className="mt-1 text-xs text-[#d74747]">{form.formState.errors.password.message}</p>
            ) : null}
          </div>

          {submitError ? <p className="text-xs text-[#d74747]">{submitError}</p> : null}

          <Button type="submit" className="h-11 w-full rounded-lg bg-[#2f66ff] text-white hover:bg-[#1f54e6]">
            Access Dashboard
          </Button>

          <p className="text-center text-sm text-[#5f6b90]">
            No account yet?{" "}
            <Link className="font-semibold text-[#2f66ff] hover:text-[#1f54e6]" href="/register">
              Create one
            </Link>
          </p>
        </form>
      </section>
    </main>
  );
}