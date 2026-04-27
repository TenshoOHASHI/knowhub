import { NextResponse } from 'next/server';
import fs from 'fs';
import path from 'path';

export async function GET() {
  const filePath = path.join(process.cwd(), 'src/lib/changelog.md');
  const md = fs.readFileSync(filePath, 'utf-8');

  const sections = md.split(/^## /m).filter(Boolean);
  const updates = sections.map((section) => {
    const firstLine = section.split('\n')[0];
    const [date, ...titleParts] = firstLine.split(' ');
    const changes = section
      .split('\n')
      .filter((line) => line.startsWith('- '))
      .map((line) => line.slice(2));
    return { date, title: titleParts.join(' '), changes };
  });

  return NextResponse.json(updates);
}
