"use client";

import { useMemo, useState } from "react";
import { Bell, LogOut, Search, Square } from "lucide-react";
import { useRouter } from "next/navigation";
import { OrdersTable } from "@/components/wms/orders-table";
import { OrderDetailDialog } from "@/components/wms/order-detail-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useWarehouseOrders } from "@/hooks/use-warehouse";
import { clearAuthSession } from "@/lib/api/auth";
import { useWmsStore } from "@/lib/store/wms-store";
import type { WmsStatus } from "@/lib/schemas";

const statusFilters: Array<{ label: string; value: WmsStatus | "ALL" }> = [
  { label: "All", value: "ALL" },
  { label: "Ready", value: "READY_TO_PICK" },
  { label: "Picking", value: "PICKING" },
  { label: "Packed", value: "PACKED" },
  { label: "Shipped", value: "SHIPPED" },
];

export default function Home() {
  const router = useRouter();
  const statusFilter = useWmsStore((state) => state.statusFilter);
  const setStatusFilter = useWmsStore((state) => state.setStatusFilter);
  const [searchTerm, setSearchTerm] = useState("");

  const logout = () => {
    clearAuthSession();
    router.push("/login");
    router.refresh();
  };

  const { data, isLoading, isFetching, error, refetch } = useWarehouseOrders(
    statusFilter === "ALL" ? undefined : statusFilter,
  );

  const totalOrders = data?.orders.length ?? 0;
  const cancelledCount = useMemo(
    () =>
      data?.orders.filter(
        (order) =>
          order.marketplace_status?.toLowerCase() === "cancelled" ||
          order.shipping_status?.toLowerCase() === "cancelled",
      ).length ?? 0,
    [data],
  );
  const shippedCount = useMemo(
    () =>
      data?.orders.filter((order) => order.wms_status === "SHIPPED").length ??
      0,
    [data],
  );

  const filteredOrders = useMemo(() => {
    const keyword = searchTerm.trim().toLowerCase();
    if (!keyword) return data?.orders ?? [];

    return (data?.orders ?? []).filter((order) => {
      return (
        order.order_sn.toLowerCase().includes(keyword) ||
        (order.marketplace_status ?? "").toLowerCase().includes(keyword) ||
        (order.shipping_status ?? "").toLowerCase().includes(keyword) ||
        (order.tracking_number ?? "").toLowerCase().includes(keyword)
      );
    });
  }, [data, searchTerm]);

  return (
    <div className="min-h-screen bg-[#f2f2f4] text-[#202020]">
      <header className="border-b border-[#0f4fff]/50 bg-[#2f66ff] text-white">
        <div className="mx-auto flex h-16 w-full max-w-[1280px] items-center justify-between px-4 sm:px-10">
          <div className="flex items-center gap-2 text-xl font-semibold">
            <Square className="h-5 w-5" />
            <span className="font-[var(--font-space-grotesk)]">WMSpaceIO</span>
          </div>

          <nav className="hidden items-center gap-6 text-sm md:flex">
            <button className="opacity-80 hover:opacity-100">Inbound</button>
            <button className="rounded-full bg-white px-5 py-1 font-semibold text-[#2f66ff]">
              Outbound
            </button>
            <button className="opacity-80 hover:opacity-100">Inventory</button>
            <button className="opacity-80 hover:opacity-100">Settings</button>
          </nav>

          <div className="flex items-center gap-3">
            <button className="rounded-full bg-white/90 p-2 text-[#2f66ff]">
              <Bell className="h-4 w-4" />
            </button>
            <div className="h-9 w-9 rounded-full border-2 border-white bg-[linear-gradient(160deg,#f8d5b4,#9c6f53)]" />
            <button
              className="rounded-full p-2 text-white/90 hover:bg-white/10"
              onClick={logout}
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>
      </header>

      <main className="mx-auto w-full  px-4 py-8 sm:px-10">
        <section className="rounded-xl    p-6">
          <div className="mb-5">
            <h1 className="font-[var(--font-space-grotesk)] text-4xl font-semibold leading-tight">
              Outbound
            </h1>
            <p className="mt-1 text-sm text-[#868696]">
              Manage all outbound process
            </p>
          </div>

          <div className="mb-5 grid gap-3 md:grid-cols-3">
            <div className="rounded-lg border border-[#dfdfe4] p-4">
              <p className="text-xs text-[#9494a0]">Total order</p>
              <p className="mt-1 text-3xl font-semibold">
                {totalOrders.toLocaleString()}
              </p>
              <p className="mt-1 text-xs text-[#2ca164]">+ 12% this month</p>
            </div>
            <div className="rounded-lg border border-[#dfdfe4] p-4">
              <p className="text-xs text-[#9494a0]">Cancelled</p>
              <p className="mt-1 text-3xl font-semibold">
                {cancelledCount.toLocaleString()}
              </p>
              <p className="mt-1 text-xs text-[#de5252]">- 5% this month</p>
            </div>
            <div className="rounded-lg border border-[#dfdfe4] p-4">
              <p className="text-xs text-[#9494a0]">Shipped</p>
              <p className="mt-1 text-3xl font-semibold">
                {shippedCount.toLocaleString()}
              </p>
              <p className="mt-1 text-xs text-[#2ca164]">+ 8% this month</p>
            </div>
          </div>

          <div className="mb-5 rounded-lg border border-[#dedee3] p-2">
            <div className="flex flex-col gap-2 md:flex-row md:items-center">
              <div className="relative md:w-[360px]">
                <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#9595a0]" />
                <Input
                  placeholder="Search here..."
                  value={searchTerm}
                  onChange={(event) => setSearchTerm(event.target.value)}
                  className="h-10 rounded-md border-[#d9d9e0] pl-9"
                />
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <Badge
                  variant="outline"
                  className="border-[#d9d9e0] bg-white text-[#525260]"
                >
                  {filteredOrders.length} Results
                </Badge>
                <Button
                  variant="outline"
                  onClick={() => refetch()}
                  disabled={isFetching}
                >
                  Refresh
                </Button>
              </div>
            </div>
          </div>

          <div className="mb-3 flex flex-wrap items-center gap-2">
            {statusFilters.map((status) => (
              <Button
                key={status.value}
                variant={statusFilter === status.value ? "secondary" : "ghost"}
                onClick={() => setStatusFilter(status.value)}
                className={
                  statusFilter === status.value
                    ? "h-8 rounded-full bg-[#2f66ff] px-4 text-white hover:bg-[#1f54e6]"
                    : "h-8 rounded-full border border-[#d8d8df] bg-white px-4 text-[#4f4f59] hover:bg-[#f2f2f7]"
                }
              >
                {status.label}
              </Button>
            ))}
          </div>

          <OrdersTable orders={filteredOrders} isLoading={isLoading} />

          {error ? (
            <p className="mt-4 text-sm text-[#de5252]">{error.message}</p>
          ) : null}
        </section>
      </main>

      <OrderDetailDialog />
    </div>
  );
}
