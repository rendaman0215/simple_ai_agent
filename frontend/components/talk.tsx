import { ScrollArea } from "@/components/ui/scroll-area";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Card } from "@/components/ui/card";
import { Sparkles } from "lucide-react";
import { cn } from "@/lib/utils";
import { Message } from "@/lib/type";
import MahjongImg from "@/app/assets/mahjong.svg";
import Image from "next/image";
import ReactMarkdown from "react-markdown";

export default function Talk({
  messages,
  isLoading,
  messagesEndRef,
}: {
  messages: Message[];
  isLoading: boolean;
  messagesEndRef: React.RefObject<HTMLDivElement>;
}) {
  return (
    <ScrollArea className="flex-1 p-4">
      <div className="space-y-4 max-w-4xl mx-auto">
        {messages.map((message) => (
          <div
            key={message.id}
            className={cn(
              "flex gap-3",
              message.sender === "user" ? "justify-end" : "justify-start"
            )}
          >
            {message.sender === "ai" && (
              <Avatar className="h-8 w-8 bg-secondary flex-shrink-0">
                <AvatarFallback className="text-secondary-foreground">
                  <Image
                    src={MahjongImg}
                    alt="Mahjong"
                    width={32}
                    height={32}
                  />
                </AvatarFallback>
              </Avatar>
            )}

            <Card
              className={cn(
                "p-3 max-w-[80%] mahjong-tile",
                message.sender === "user"
                  ? "bg-primary text-muted-foreground"
                  : "bg-card text-card-foreground"
              )}
            >
              <p className="text-sm leading-relaxed">
                <ReactMarkdown>{message.content}</ReactMarkdown>
              </p>
              <p
                className={cn(
                  "text-xs mt-2 opacity-70",
                  message.sender === "user"
                    ? "text-muted-foreground"
                    : "text-muted-foreground"
                )}
              >
                {message.timestamp.toLocaleTimeString("ja-JP", {
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </p>
            </Card>

            {message.sender === "user" && (
              <Avatar className="h-8 w-8 bg-muted flex-shrink-0">
                <AvatarFallback className="text-muted-foreground">
                  あ
                </AvatarFallback>
              </Avatar>
            )}
          </div>
        ))}

        {isLoading && (
          <div className="flex gap-3 justify-start">
            <Avatar className="h-8 w-8 bg-secondary flex-shrink-0">
              <AvatarFallback className="text-secondary-foreground">
                <Sparkles className="h-4 w-4" />
              </AvatarFallback>
            </Avatar>
            <Card className="p-3 bg-card text-card-foreground mahjong-tile">
              <div className="flex gap-1">
                <div className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce" />
                <div
                  className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce"
                  style={{ animationDelay: "0.1s" }}
                />
                <div
                  className="w-2 h-2 bg-muted-foreground rounded-full animate-bounce"
                  style={{ animationDelay: "0.2s" }}
                />
              </div>
            </Card>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>
    </ScrollArea>
  );
}
