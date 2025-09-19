import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import Image from "next/image";
import MahjongImg from "@/app/assets/mahjong.svg";

export default function Header() {
  return (
    <div className="bg-secondary border-b border-border p-4">
      <div className="flex items-center gap-3">
        <Avatar className="h-8 w-8">
          <AvatarFallback>
            <Image src={MahjongImg} alt="Mahjong" width={32} height={32} />
          </AvatarFallback>
        </Avatar>
        <div>
          <h1 className="font-semibold text-foreground">麻雀AIエージェント</h1>
          <p className="text-xs text-muted-foreground">オンライン</p>
        </div>
      </div>
    </div>
  );
}
