'use client';

import { useSidebar } from '@/context/SidebarContext';
import { FiFolder } from 'react-icons/fi';

const CATEGORIES = [
  { name: 'Go', count: 0 },
  { name: 'gRPC', count: 1 },
  { name: 'Database', count: 2 },
  { name: 'DevOps', count: 3 },
  { name: 'Docker', count: 4 },
  { name: 'TypeScript', count: 0 },
  { name: 'React', count: 1 },
  { name: 'Next.js', count: 2 },
  { name: 'Authentication', count: 3 },
  { name: 'Testing', count: 4 },
  { name: 'Architecture', count: 0 },
  { name: 'CI/CD', count: 1 },
  { name: 'Security', count: 2 },
  { name: 'Networking', count: 3 },
  { name: 'OS', count: 4 },
  { name: 'Algorithms', count: 0 },
  { name: 'Design Patterns', count: 1 },
  { name: 'Git', count: 2 },
  { name: 'Linux', count: 3 },
  { name: 'Kubernetes', count: 4 },
  { name: 'AWS', count: 0 },
  { name: 'GCP', count: 1 },
  { name: 'Azure', count: 2 },
  { name: 'Terraform', count: 3 },
  { name: 'Ansible', count: 4 },
  { name: 'GraphQL', count: 0 },
  { name: 'REST API', count: 1 },
  { name: 'WebSocket', count: 2 },
  { name: 'Microservices', count: 3 },
  { name: 'Monolith', count: 4 },
  { name: 'Event Driven', count: 0 },
  { name: 'Domain Driven Design', count: 1 },
  { name: 'Clean Architecture', count: 2 },
  { name: 'Hexagonal', count: 3 },
  { name: 'CQRS', count: 4 },
  { name: 'Event Sourcing', count: 0 },
  { name: 'Redis', count: 1 },
  { name: 'Message Queue', count: 2 },
  { name: 'Kafka', count: 3 },
  { name: 'RabbitMQ', count: 4 },
  { name: 'PostgreSQL', count: 0 },
  { name: 'MongoDB', count: 1 },
  { name: 'SQLite', count: 2 },
  { name: 'Elasticsearch', count: 3 },
  { name: 'Nginx', count: 4 },
  { name: 'Apache', count: 0 },
  { name: 'Caddy', count: 1 },
  { name: 'Prometheus', count: 2 },
  { name: 'Grafana', count: 3 },
  { name: 'Logging', count: 4 },
  { name: 'Tracing', count: 0 },
  { name: 'Observability', count: 1 },
  { name: 'SRE', count: 2 },
  { name: 'DevSecOps', count: 3 },
  { name: 'Agile', count: 4 },
  { name: 'Scrum', count: 0 },
  { name: 'Kanban', count: 1 },
  { name: 'Code Review', count: 2 },
  { name: 'Pair Programming', count: 3 },
  { name: 'TDD', count: 4 },
  { name: 'BDD', count: 0 },
  { name: 'DDD', count: 1 },
  { name: 'SOLID', count: 2 },
  { name: 'DRY', count: 3 },
  { name: 'KISS', count: 4 },
  { name: 'YAGNI', count: 0 },
  { name: 'Design Principles', count: 1 },
  { name: 'Refactoring', count: 2 },
  { name: 'Performance', count: 3 },
  { name: 'Scalability', count: 4 },
  { name: 'Reliability', count: 0 },
  { name: 'Monitoring', count: 1 },
  { name: 'Alerting', count: 2 },
  { name: 'Incident Response', count: 3 },
  { name: 'Post Mortem', count: 4 },
  { name: 'On Call', count: 0 },
  { name: 'Runbooks', count: 1 },
  { name: 'Documentation', count: 2 },
  { name: 'API Design', count: 3 },
  { name: 'Versioning', count: 4 },
  { name: 'Dependency Management', count: 0 },
  { name: 'Build Tools', count: 1 },
  { name: 'Webpack', count: 2 },
  { name: 'Vite', count: 3 },
  { name: 'ESBuild', count: 4 },
  { name: 'Turbopack', count: 0 },
  { name: 'Bun', count: 1 },
  { name: 'Deno', count: 2 },
  { name: 'Rust', count: 3 },
  { name: 'Python', count: 4 },
  { name: 'Java', count: 0 },
  { name: 'CSharp', count: 1 },
  { name: 'Swift', count: 2 },
  { name: 'Kotlin', count: 3 },
  { name: 'Flutter', count: 4 },
  { name: 'iOS', count: 0 },
  { name: 'Android', count: 1 },
  { name: 'Mobile', count: 2 },
  { name: 'PWA', count: 3 },
  { name: 'Accessibility', count: 4 },
  { name: 'I18n', count: 0 },
  { name: 'SEO', count: 1 },
  { name: 'Web Performance', count: 2 },
  { name: 'Core Web Vitals', count: 3 },
  { name: 'Browser', count: 4 },
  { name: 'HTTP', count: 0 },
  { name: 'TCPIP', count: 1 },
  { name: 'DNS', count: 2 },
] as const;

export default function Sidebar() {
  const { isOpen } = useSidebar();

  if (!isOpen) return null;

  return (
    <aside className='w-48 border-r border-black dark:border-stone-600 shrink-0 overflow-y-auto h-full thin-scrollbar'>
      <h2 className='font-semibold mb-3 text-md sticky top-0 bg-white dark:bg-stone-900/90 py-4 px-4 z-10'>
        カテゴリ
      </h2>
      <ul className='space-y-2 px-4 pb-4'>
        {CATEGORIES.map((cat) => (
          <li key={cat.name}>
            <button className='flex items-center gap-2 w-full text-left text-sm text-gray-600 dark:text-stone-400 hover:text-black dark:hover:text-stone-100'>
              <FiFolder className='shrink-0' />
              <span>{cat.name}</span>
              <span className='text-xs text-gray-400 dark:text-stone-500'>
                ({cat.count})
              </span>
            </button>
          </li>
        ))}
      </ul>
    </aside>
  );
}
