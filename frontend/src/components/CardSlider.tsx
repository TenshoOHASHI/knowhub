import { FiChevronLeft, FiChevronRight } from 'react-icons/fi';
import { PortfolioCard } from './PortfolioCard';
import { PortfolioItem } from '@/lib/types';
import { useRef } from 'react';

export function CardSlider({ items }: { items: PortfolioItem[] }) {
  const scrollRef = useRef<HTMLDivElement>(null);

  const scroll = (direction: 'left' | 'right') => {
    if (!scrollRef.current) return;
    const cardWidth = scrollRef.current.offsetWidth / 3;
    scrollRef.current.scrollBy({
      left: direction === 'left' ? -cardWidth : cardWidth,
      behavior: 'smooth',
    });
  };

  const canScroll = items.length > 3;

  return (
    <div className='relative group/slider'>
      {/* Scroll container */}
      <div
        ref={scrollRef}
        className='flex gap-4 overflow-x-auto scroll-smooth snap-x snap-mandatory py-4 pb-2'
        style={{ scrollbarWidth: 'none', msOverflowStyle: 'none' }}
      >
        {items.map((item, index) => (
          <div
            key={item.id}
            className='snap-start shrink-0 w-full md:w-[calc(50%-8px)] lg:w-[calc(33.333%-11px)]'
          >
            <PortfolioCard item={item} index={index} />
          </div>
        ))}
      </div>

      {/* Nav buttons - only when 4+ items */}
      {canScroll && (
        <>
          <button
            onClick={() => scroll('left')}
            className='absolute left-0 top-1/2 -translate-y-1/2 -translate-x-3 w-9 h-9 rounded-full bg-white dark:bg-stone-800 border border-stone-200 dark:border-stone-700 shadow-md flex items-center justify-center opacity-0 group-hover/slider:opacity-100 transition-opacity hover:bg-stone-50 dark:hover:bg-stone-700'
          >
            <FiChevronLeft size={18} />
          </button>
          <button
            onClick={() => scroll('right')}
            className='absolute right-0 top-1/2 -translate-y-1/2 translate-x-3 w-9 h-9 rounded-full bg-white dark:bg-stone-800 border border-stone-200 dark:border-stone-700 shadow-md flex items-center justify-center opacity-0 group-hover/slider:opacity-100 transition-opacity hover:bg-stone-50 dark:hover:bg-stone-700'
          >
            <FiChevronRight size={18} />
          </button>
        </>
      )}
    </div>
  );
}
