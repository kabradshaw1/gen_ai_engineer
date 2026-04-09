import { JavaSubHeader } from "@/components/java/JavaSubHeader";

export default function JavaTasksLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <JavaSubHeader />
      {children}
    </>
  );
}
