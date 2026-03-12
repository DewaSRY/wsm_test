import { create } from "zustand";
import type { WmsStatus } from "@/lib/schemas";

type StatusFilter = WmsStatus | "ALL";

type WmsStore = {
  statusFilter: StatusFilter;
  selectedOrderSn: string | null;
  detailOpen: boolean;
  setStatusFilter: (status: StatusFilter) => void;
  openOrderDetail: (orderSn: string) => void;
  closeOrderDetail: () => void;
};

export const useWmsStore = create<WmsStore>((set) => ({
  statusFilter: "ALL",
  selectedOrderSn: null,
  detailOpen: false,
  setStatusFilter: (statusFilter) => set({ statusFilter }),
  openOrderDetail: (selectedOrderSn) => set({ selectedOrderSn, detailOpen: true }),
  closeOrderDetail: () => set({ detailOpen: false, selectedOrderSn: null }),
}));