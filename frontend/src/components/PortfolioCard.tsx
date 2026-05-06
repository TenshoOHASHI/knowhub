import { PortfolioItem } from '@/lib/types';
import { motion } from 'motion/react';
import { FiCheckCircle, FiCode, FiCalendar, FiExternalLink } from 'react-icons/fi';
import { useMemo } from 'react';

export function PortfolioCard({
  item,
  index,
}: {
  item: PortfolioItem;
  index: number;
}) {
  const isCompleted = item.status === 'completed';

  const categoryMap: Record<string, string> = {
    project: 'Project',
    assignment: 'Assignment',
    contribution: 'Contribution',
  };
  const categoryLabel = categoryMap[item.category] || 'Project';

  const techStack = useMemo(() => {
    try {
      const parsed = JSON.parse(item.tech_stack || '[]');
      return Array.isArray(parsed) ? parsed : [];
    } catch {
      return [];
    }
  }, [item.tech_stack]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 15 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ delay: index * 0.1, duration: 0.4 }}
      whileHover={{ y: -4 }}
      className='group relative rounded-xl border border-stone-200 dark:border-stone-700 bg-white dark:bg-stone-900 p-5 transition-shadow hover:shadow-lg h-full flex flex-col'
    >
      {/* Badges */}
      <div className='flex items-center gap-2 mb-3'>
        <span
          className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full ${
            isCompleted
              ? 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400'
              : 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
          }`}
        >
          {isCompleted ? <FiCheckCircle size={10} /> : <FiCode size={10} />}
          {isCompleted ? 'Completed' : 'In Progress'}
        </span>
        <span className='text-xs text-gray-400 bg-stone-100 dark:bg-stone-800 px-2 py-0.5 rounded-full'>
          {categoryLabel}
        </span>
      </div>

      {/* Title */}
      <h3 className='text-lg font-semibold mb-1 group-hover:text-green-700 dark:group-hover:text-green-400 transition-colors'>
        {item.title}
      </h3>

      {/* Description */}
      <p className='text-sm text-gray-500 dark:text-stone-400 line-clamp-3 mb-3'>
        {item.description}
      </p>

      {/* Tech Stack */}
      {techStack.length > 0 && (
        <div className='flex flex-wrap gap-1 mb-3'>
          {techStack.map((tech: string) => (
            <span
              key={tech}
              className='text-xs px-2 py-0.5 rounded bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400'
            >
              {tech}
            </span>
          ))}
        </div>
      )}

      {/* Date */}
      {item.created_at?.seconds && (
        <p className='text-xs text-stone-400 dark:text-stone-500 flex items-center gap-1'>
          <FiCalendar size={12} />
          {new Date(item.created_at.seconds * 1000).toLocaleDateString('ja-JP', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
          })}
        </p>
      )}

      {/* Link */}
      {item.url && (
        <a
          href={item.url}
          target='_blank'
          rel='noopener noreferrer'
          className='mt-auto inline-flex items-center gap-1 text-sm text-gray-400 hover:text-black dark:hover:text-white transition-colors'
        >
          <FiExternalLink size={14} />
          View project
        </a>
      )}
    </motion.div>
  );
}
