import Sidebar from '@/components/Sidebar';

export default function WikiLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <style>{`footer { display: none !important; }`}</style>
      <div className='flex flex-1 overflow-hidden h-full'>
        <Sidebar />
        <div className='flex-1 overflow-y-auto'>{children}</div>
      </div>
    </>
  );
}
