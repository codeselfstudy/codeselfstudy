"use client";

import * as React from "react";
import { Menu as MenuPrimitive } from "@base-ui/react/menu";

import { cn } from "@/lib/utils";

// House-style dropdown menu over @base-ui/react's Menu primitive, matching the
// Sheet wrapper in ui/sheet.tsx (base-vega style, see components.json). The
// primitive owns keyboard navigation, focus management and dismiss-on-outside /
// Escape, so consumers only supply content. Written by hand rather than via the
// shadcn CLI, whose default registry emits Radix, not the @base-ui primitives
// this codebase standardizes on.

const itemClass =
  "data-highlighted:bg-accent data-highlighted:text-accent-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm no-underline outline-none select-none hover:no-underline data-disabled:pointer-events-none data-disabled:opacity-50";

function DropdownMenu({ ...props }: MenuPrimitive.Root.Props) {
  return <MenuPrimitive.Root data-slot="dropdown-menu" {...props} />;
}

function DropdownMenuTrigger({ ...props }: MenuPrimitive.Trigger.Props) {
  return <MenuPrimitive.Trigger data-slot="dropdown-menu-trigger" {...props} />;
}

function DropdownMenuContent({
  className,
  sideOffset = 8,
  align = "end",
  children,
  ...props
}: MenuPrimitive.Popup.Props & {
  sideOffset?: number;
  align?: "start" | "center" | "end";
}) {
  return (
    <MenuPrimitive.Portal>
      <MenuPrimitive.Positioner
        data-slot="dropdown-menu-positioner"
        className="z-50"
        sideOffset={sideOffset}
        align={align}
      >
        <MenuPrimitive.Popup
          data-slot="dropdown-menu-content"
          className={cn(
            "bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 min-w-56 rounded-md border p-1 shadow-md outline-none",
            className
          )}
          {...props}
        >
          {children}
        </MenuPrimitive.Popup>
      </MenuPrimitive.Positioner>
    </MenuPrimitive.Portal>
  );
}

function DropdownMenuItem({ className, ...props }: MenuPrimitive.Item.Props) {
  return (
    <MenuPrimitive.Item
      data-slot="dropdown-menu-item"
      className={cn(itemClass, className)}
      {...props}
    />
  );
}

// DropdownMenuLinkItem is a menu item that navigates — it renders an <a>, so it
// gets native link semantics (open in new tab, copy link) on top of the menu
// item's keyboard handling.
function DropdownMenuLinkItem({
  className,
  ...props
}: MenuPrimitive.LinkItem.Props) {
  return (
    <MenuPrimitive.LinkItem
      data-slot="dropdown-menu-link-item"
      className={cn(itemClass, className)}
      {...props}
    />
  );
}

// DropdownMenuLabel is a non-interactive header (e.g. the signed-in username).
function DropdownMenuLabel({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="dropdown-menu-label"
      className={cn(
        "text-muted-foreground px-2 py-1.5 text-xs font-medium",
        className
      )}
      {...props}
    />
  );
}

function DropdownMenuSeparator({
  className,
  ...props
}: MenuPrimitive.Separator.Props) {
  return (
    <MenuPrimitive.Separator
      data-slot="dropdown-menu-separator"
      className={cn("bg-border -mx-1 my-1 h-px", className)}
      {...props}
    />
  );
}

export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLinkItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
};
