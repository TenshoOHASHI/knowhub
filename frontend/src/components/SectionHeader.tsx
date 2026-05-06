import { motion } from 'motion/react';
export function SectionHeader({
  icon,
  bgColor,
  label,
  count,
}: {
  icon: React.ReactNode;
  bgColor: string;
  label: string;
  count: number;
}) {
  return (
    <motion.h2
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true }}
      className='flex items-center gap-2 text-lg font-semibold mb-6'
    >
      <span
        className={`flex items-center justify-center w-8 h-8 rounded-lg ${bgColor}`}
      >
        {icon}
      </span>
      {label}
      <span className='text-sm font-normal text-gray-400 ml-1'>({count})</span>
    </motion.h2>
  );
}
