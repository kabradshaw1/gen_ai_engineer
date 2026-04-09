import { GoSubHeader } from "@/components/go/GoSubHeader";
import { AiAssistantDrawer } from "@/components/go/AiAssistantDrawer";

export default function GoEcommerceLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <GoSubHeader />
      {children}
      <AiAssistantDrawer />
    </>
  );
}
