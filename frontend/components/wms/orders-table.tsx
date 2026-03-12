"use client";

import { useEffect, useMemo, useState } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { ChevronsUpDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useWmsStore } from "@/lib/store/wms-store";
import { formatDateTime } from "@/lib/utils";
import type { OrderListItem } from "@/lib/schemas";

type OrdersTableProps = {
  orders: OrderListItem[];
  isLoading: boolean;
};

const columnHelper = createColumnHelper<OrderListItem>();

function statusClass(value: string | null | undefined, type: "marketplace" | "shipping" | "wms") {
  const status = (value ?? "-").toLowerCase();

  if (type === "wms") {
    if (status.includes("ready")) return "bg-[#fff4e8] text-[#db8d3f]";
    if (status.includes("pick")) return "bg-[#fff8df] text-[#b88220]";
    if (status.includes("pack")) return "bg-[#efe9ff] text-[#6b4acc]";
    if (status.includes("ship")) return "bg-[#e9f8ef] text-[#2e9c65]";
    return "bg-[#f0f0f3] text-[#6d6d78]";
  }

  if (status.includes("cancel")) return "bg-[#ffecec] text-[#db5a5a]";
  if (status.includes("deliver")) return "bg-[#e7f9ee] text-[#38a76f]";
  if (status.includes("ship") || status.includes("approved")) return "bg-[#fff4e8] text-[#db8d3f]";
  if (status.includes("await") || status.includes("label") || status.includes("process")) {
    return "bg-[#ecefff] text-[#6b76cf]";
  }
  return "bg-[#f0f0f3] text-[#6d6d78]";
}

type HeaderLabelProps = {
  label: string;
  withIcon?: boolean;
};

function HeaderLabel({ label, withIcon = true }: HeaderLabelProps) {
  return (
    <span className="inline-flex items-center gap-1.5 whitespace-nowrap">
      {label}
      {withIcon ? <ChevronsUpDown className="h-3 w-3 text-[#8f8f9e]" /> : null}
    </span>
  );
}

export function OrdersTable({ orders, isLoading }: OrdersTableProps) {
  const openOrderDetail = useWmsStore((state) => state.openOrderDetail);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  useEffect(() => {
    const maxPage = Math.max(1, Math.ceil(orders.length / pageSize));
    if (page > maxPage) {
      setPage(maxPage);
    }
  }, [orders.length, page, pageSize]);

  const pagedOrders = useMemo(() => {
    const start = (page - 1) * pageSize;
    return orders.slice(start, start + pageSize);
  }, [orders, page, pageSize]);

  const totalPages = Math.max(1, Math.ceil(orders.length / pageSize));

  const visiblePages = useMemo(() => {
    if (totalPages <= 5) return Array.from({ length: totalPages }, (_, index) => index + 1);
    if (page <= 3) return [1, 2, 3, 4, 5];
    if (page >= totalPages - 2) return [totalPages - 4, totalPages - 3, totalPages - 2, totalPages - 1, totalPages];
    return [page - 2, page - 1, page, page + 1, page + 2];
  }, [page, totalPages]);

  const columns = [
    columnHelper.accessor("order_sn", {
      header: () => <HeaderLabel label="Order SN" />,
      cell: (info) => (
        <button className="font-medium text-[#353546] hover:text-[#2f66ff]" onClick={() => openOrderDetail(info.getValue())}>
          {info.getValue()}
        </button>
      ),
    }),
    columnHelper.accessor("marketplace_status", {
      header: () => <HeaderLabel label="Marketplace Status" />,
      cell: (info) => (
        <span
          className={`inline-flex rounded-md px-2 py-1 text-xs font-medium ${statusClass(info.getValue(), "marketplace")}`}
        >
          {(info.getValue() ?? "-").replaceAll("_", " ")}
        </span>
      ),
    }),
    columnHelper.accessor("shipping_status", {
      header: () => <HeaderLabel label="Shipping Status" />,
      cell: (info) => (
        <span className={`inline-flex rounded-md px-2 py-1 text-xs font-medium ${statusClass(info.getValue(), "shipping")}`}>
          {(info.getValue() ?? "-").replaceAll("_", " ")}
        </span>
      ),
    }),
    columnHelper.accessor("wms_status", {
      header: () => <HeaderLabel label="WMS Status" />,
      cell: (info) => (
        <span className={`inline-flex rounded-md px-2 py-1 text-xs font-medium ${statusClass(info.getValue(), "wms")}`}>
          {info.getValue().replaceAll("_", " ")}
        </span>
      ),
    }),
    columnHelper.accessor("tracking_number", {
      header: () => <HeaderLabel label="Tracking Number" />,
      cell: (info) => info.getValue() ?? "-",
    }),
    columnHelper.accessor("updated_at", {
      header: () => <HeaderLabel label="Update At" />,
      cell: (info) => formatDateTime(info.getValue()),
    }),
    columnHelper.display({
      id: "actions",
      header: () => <HeaderLabel label="Action" withIcon={false} />,
      cell: ({ row }) => {
        return (
          <Button
            size="sm"
            className="h-7 rounded-md bg-[#2f66ff] px-3 text-xs text-white hover:bg-[#1f54e6]"
            onClick={() => openOrderDetail(row.original.order_sn)}
          >
            Detail
          </Button>
        );
      },
    }),
  ];

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data: pagedOrders,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  if (isLoading) {
    return <p className="py-8 text-sm text-[#6a6a77]">Loading orders...</p>;
  }

  if (!orders.length) {
    return <p className="py-8 text-sm text-[#6a6a77]">No orders found for this filter.</p>;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-[#dadadf] bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-full border-collapse text-xs">
          <thead className="bg-[#f7f7fb]">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th key={header.id} className="border-b border-[#e7e7ee] px-3 py-3 text-left font-medium text-[#5f5f6d]">
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row, index) => (
              <tr
                key={row.id}
                className={`border-t border-[#ececf2] ${index % 2 === 0 ? "bg-white" : "bg-[#fbfbfe]"} hover:bg-[#f1f4ff]`}
              >
                {row.getVisibleCells().map((cell) => (
                  <td key={cell.id} className="px-3 py-2.5 align-middle text-[#363646]">
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex flex-col items-start justify-between gap-3 border-t border-[#e5e5ea] bg-white px-3 py-2.5 text-xs text-[#5f5f6d] md:flex-row md:items-center">
        <div className="flex items-center gap-1.5">
          <span>Show</span>
          <select
            className="h-6 rounded border border-[#d8d8df] bg-white px-1"
            value={pageSize}
            onChange={(event) => {
              const nextSize = Number(event.target.value);
              setPageSize(nextSize);
              setPage(1);
            }}
          >
            <option value={10}>10</option>
            <option value={25}>25</option>
            <option value={50}>50</option>
          </select>
          <span>of {orders.length} entries</span>
        </div>

        <div className="flex items-center gap-1">
          <button
            className="h-6 min-w-6 rounded border border-[#d8d8df] bg-white px-1 disabled:opacity-40"
            disabled={page <= 1}
            onClick={() => setPage((prev) => Math.max(prev - 1, 1))}
          >
            &lt;
          </button>
          {visiblePages.map((pageNumber) => (
            <button
              key={pageNumber}
              className={`h-6 min-w-6 rounded border px-2 ${
                pageNumber === page
                  ? "border-[#2f66ff] bg-[#eef2ff] text-[#2f66ff]"
                  : "border-[#d8d8df] bg-white text-[#5f5f6d]"
              }`}
              onClick={() => setPage(pageNumber)}
            >
              {pageNumber}
            </button>
          ))}
          <button
            className="h-6 min-w-6 rounded border border-[#d8d8df] bg-white px-1 disabled:opacity-40"
            disabled={page >= totalPages}
            onClick={() => setPage((prev) => Math.min(prev + 1, totalPages))}
          >
            &gt;
          </button>
        </div>
      </div>
    </div>
  );
}