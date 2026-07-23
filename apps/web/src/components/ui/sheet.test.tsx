import { describe, expect, test } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";

// The navbar drawer only uses Sheet/Trigger/Content/Title, so the header,
// footer, description and standalone close parts are exercised here instead.
// The jsdom stubs Base UI's Dialog needs live in `test/setup.ts`.
function SheetHarness({
  side,
  showCloseButton,
}: {
  side?: "top" | "right" | "bottom" | "left";
  showCloseButton?: boolean;
}) {
  return (
    <Sheet>
      <SheetTrigger>Open panel</SheetTrigger>
      <SheetContent side={side} showCloseButton={showCloseButton}>
        <SheetHeader>
          <SheetTitle>Panel title</SheetTitle>
          <SheetDescription>Panel description</SheetDescription>
        </SheetHeader>
        <SheetFooter>
          {/* Not named "Close": SheetContent's built-in close button already
              uses that accessible name, and two would make the query ambiguous. */}
          <SheetClose>Done</SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}

describe("Sheet", () => {
  test("renders the header, title, description and footer once opened", async () => {
    const user = userEvent.setup();
    render(<SheetHarness />);

    expect(screen.queryByText("Panel title")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Open panel" }));

    const dialog = await screen.findByRole("dialog");
    expect(screen.getByText("Panel title")).toBeInTheDocument();
    expect(screen.getByText("Panel description")).toBeInTheDocument();
    expect(dialog.querySelector("[data-slot='sheet-header']")).not.toBeNull();
    expect(dialog.querySelector("[data-slot='sheet-footer']")).not.toBeNull();
  });

  test("closes when a SheetClose control is clicked", async () => {
    const user = userEvent.setup();
    render(<SheetHarness />);

    await user.click(screen.getByRole("button", { name: "Open panel" }));
    await screen.findByRole("dialog");

    await user.click(screen.getByRole("button", { name: "Done" }));

    await waitFor(() =>
      expect(screen.queryByText("Panel title")).not.toBeInTheDocument()
    );
  });

  test("defaults to the right side and honours an explicit side", async () => {
    const user = userEvent.setup();
    const { unmount } = render(<SheetHarness />);

    await user.click(screen.getByRole("button", { name: "Open panel" }));
    expect(await screen.findByRole("dialog")).toHaveAttribute(
      "data-side",
      "right"
    );

    unmount();
    render(<SheetHarness side="left" />);

    await user.click(screen.getByRole("button", { name: "Open panel" }));
    expect(await screen.findByRole("dialog")).toHaveAttribute(
      "data-side",
      "left"
    );
  });

  test("omits the built-in close button when showCloseButton is false", async () => {
    const user = userEvent.setup();
    render(<SheetHarness showCloseButton={false} />);

    await user.click(screen.getByRole("button", { name: "Open panel" }));
    await screen.findByRole("dialog");

    expect(
      screen.queryByRole("button", { name: "Close" })
    ).not.toBeInTheDocument();
    // The caller's own close control is unaffected.
    expect(screen.getByRole("button", { name: "Done" })).toBeInTheDocument();
  });
});
