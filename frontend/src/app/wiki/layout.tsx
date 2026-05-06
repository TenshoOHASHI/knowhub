import Sidebar from '@/components/Sidebar';
import { Suspense } from 'react';

export default function WikiLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <style>{`footer { display: none !important; }`}</style>
      <div className='flex flex-1 overflow-hidden h-full'>
        <Suspense fallback={null}>
          <Sidebar />
        </Suspense>
        <div className='flex-1 overflow-y-auto'>
          <Suspense fallback={null}>{children}</Suspense>
        </div>
      </div>
    </>
  );
}
