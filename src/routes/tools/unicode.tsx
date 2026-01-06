import { PageWrapper } from "@/components/PageWrapper";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { sansSerif } from "@/data/unicode";

export const Route = createFileRoute("/tools/unicode")({
  component: UnicodeTool,
});

function UnicodeTool() {
  const [text, setText] = useState("");
  const [output, setOutput] = useState("");

  function handleOnKeyUp(e: React.ChangeEvent<HTMLInputElement>) {
    const chars = e.target.value;
    const translated = [...chars]
      .map((char) => {
        return sansSerif[char] || char;
      })
      .join("");
    setText(chars);
    setOutput(translated);
  }

  return (
    <PageWrapper>
      <Card className="mx-auto max-w-2xl">
        <CardHeader>
          <CardTitle className="text-3xl font-bold">
            Bold Unicode Text Tool
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground mb-4">
            Type your text below to get the unicode characters.
          </p>
          <Input
            type="text"
            placeholder="Enter some text"
            onChange={handleOnKeyUp}
            className="mb-6 font-mono"
          />
          {text && (
            <div className="rounded-md border bg-gray-50 p-4 dark:bg-gray-900">
              <div className="text-muted-foreground mb-1 font-mono text-xs tracking-wider uppercase">
                Output:
              </div>
              <div className="text-xl break-words">{output}</div>
            </div>
          )}
        </CardContent>
      </Card>
    </PageWrapper>
  );
}
