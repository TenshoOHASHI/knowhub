'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import * as d3 from 'd3';
import { getKnowledgeGraph, type EntityNode, type RelationEdge } from '@/lib/api';
import Link from 'next/link';

interface SimNode extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: string;
  articleIds: string[];
}

interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  label: string;
}

const TYPE_COLORS: Record<string, string> = {
  Technology: '#3b82f6',
  Protocol: '#8b5cf6',
  Company: '#f59e0b',
  Concept: '#10b981',
  Language: '#ef4444',
  Tool: '#06b6d4',
  Framework: '#ec4899',
  Library: '#f97316',
  Platform: '#14b8a6',
  Method: '#a855f7',
};

function nodeColor(type: string): string {
  return TYPE_COLORS[type] ?? '#6b7280';
}

export default function KnowledgeGraph() {
  const svgRef = useRef<SVGSVGElement>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState<EntityNode | null>(null);
  const [graphData, setGraphData] = useState<{
    entities: EntityNode[];
    relations: RelationEdge[];
  } | null>(null);

  const fetchGraph = useCallback(() => {
    setLoading(true);
    setError(null);
    setGraphData(null);

    let cancelled = false;

    getKnowledgeGraph()
      .then((data) => {
        if (cancelled) return;
        // まず loading を false にして SVG を DOM にマウントさせる
        // その後 graphData をセット → 別の useEffect で描画
        setLoading(false);
        setGraphData(data);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err.message);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  // 初回フェッチ
  useEffect(() => {
    const cancel = fetchGraph();
    return cancel;
  }, [fetchGraph]);

  // graphData がセットされた後（SVG が DOM にある状態）で D3 描画
  useEffect(() => {
    if (!graphData || !svgRef.current) return;

    const entities = graphData.entities;
    const relations = graphData.relations;
    if (entities.length === 0) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    const width = svgRef.current.clientWidth || 800;
    const height = svgRef.current.clientHeight || 500;

    const g = svg.append('g');

    // Zoom
    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.2, 5])
      .on('zoom', (event) => g.attr('transform', event.transform));
    svg.call(zoom);

    const nodes: SimNode[] = entities.map((e) => ({
      id: e.id,
      name: e.name,
      type: e.type,
      articleIds: e.article_ids ?? [],
    }));

    const idSet = new Set(nodes.map((n) => n.id));
    const links: SimLink[] = relations
      .filter((r) => idSet.has(r.source) && idSet.has(r.target))
      .map((r) => ({
        source: r.source,
        target: r.target,
        label: r.label,
      }));

    const maxArticles = Math.max(1, ...nodes.map((n) => n.articleIds.length));

    // ノード半径: 記事数に応じて 16〜40px
    const nodeRadius = (d: SimNode) =>
      16 + (d.articleIds.length / maxArticles) * 24;

    const simulation = d3
      .forceSimulation<SimNode>(nodes)
      .force(
        'link',
        d3
          .forceLink<SimNode, SimLink>(links)
          .id((d) => d.id)
          .distance(200),
      )
      .force('charge', d3.forceManyBody().strength(-500))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force(
        'collision',
        d3.forceCollide<SimNode>().radius((d) => nodeRadius(d) + 8),
      );

    // Links
    const link = g
      .append('g')
      .selectAll('line')
      .data(links)
      .join('line')
      .attr('stroke', '#64748b')
      .attr('stroke-opacity', 0.4)
      .attr('stroke-width', 2);

    // Link labels — 背景付きラベルで視認性向上
    const linkLabelGroup = g
      .append('g')
      .selectAll<SVGGElement, SimLink>('g')
      .data(links)
      .join('g');

    // ラベル背景
    linkLabelGroup
      .append('rect')
      .attr('rx', 4)
      .attr('ry', 4)
      .attr('fill', 'rgba(30,41,59,0.85)')
      .attr('stroke', '#475569')
      .attr('stroke-width', 0.5);

    // ラベルテキスト
    const linkLabel = linkLabelGroup
      .append('text')
      .text((d) => d.label)
      .attr('font-size', 13)
      .attr('font-weight', 500)
      .attr('fill', '#cbd5e1')
      .attr('text-anchor', 'middle')
      .attr('dy', 4)
      .style('pointer-events', 'none');

    // テキスト描画後に背景 rect をテキスト幅に合わせる
    linkLabel.each(function (d) {
      const textEl = this as SVGTextElement;
      const bbox = textEl.getBBox();
      const parent = textEl.parentNode as SVGGElement;
      const rect = parent.querySelector('rect');
      if (rect) {
        d3.select(rect)
          .attr('x', bbox.x - 6)
          .attr('y', bbox.y - 3)
          .attr('width', bbox.width + 12)
          .attr('height', bbox.height + 6);
      }
      void d;
    });

    // Node groups
    const node = g
      .append('g')
      .selectAll<SVGGElement, SimNode>('g')
      .data(nodes)
      .join('g')
      .call(
        d3
          .drag<SVGGElement, SimNode>()
          .on('start', (event, d) => {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on('drag', (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on('end', (event, d) => {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          }),
      );

    // Circles — ドロップシャドウ付き
    const shadow = svg
      .append('defs')
      .append('filter')
      .attr('id', 'shadow')
      .attr('x', '-50%')
      .attr('y', '-50%')
      .attr('width', '200%')
      .attr('height', '200%');
    shadow
      .append('feDropShadow')
      .attr('dx', 0)
      .attr('dy', 2)
      .attr('stdDeviation', 3)
      .attr('flood-color', 'rgba(0,0,0,0.4)');

    node
      .append('circle')
      .attr('r', nodeRadius)
      .attr('fill', (d) => nodeColor(d.type))
      .attr('stroke', '#fff')
      .attr('stroke-width', 2.5)
      .attr('filter', 'url(#shadow)')
      .style('cursor', 'pointer')
      .on('mouseenter', function () {
        d3.select(this)
          .transition()
          .duration(150)
          .attr('stroke-width', 4)
          .attr('stroke', '#fbbf24');
      })
      .on('mouseleave', function () {
        d3.select(this)
          .transition()
          .duration(150)
          .attr('stroke-width', 2.5)
          .attr('stroke', '#fff');
      })
      .on('click', (_event, d) => {
        setSelected({
          id: d.id,
          name: d.name,
          type: d.type,
          article_ids: d.articleIds,
        });
      });

    // Labels — 常時表示
    node
      .append('text')
      .text((d) => d.name)
      .attr('font-size', 14)
      .attr('font-weight', 600)
      .attr('text-anchor', 'middle')
      .attr('dy', (d) => nodeRadius(d) + 18)
      .attr('fill', '#e2e8f0')
      .style('pointer-events', 'none')
      .style('text-shadow', '0 1px 4px rgba(0,0,0,0.9), 0 0 8px rgba(0,0,0,0.6)');

    // Tick
    simulation.on('tick', () => {
      link
        .attr('x1', (d) => (d.source as SimNode).x!)
        .attr('y1', (d) => (d.source as SimNode).y!)
        .attr('x2', (d) => (d.target as SimNode).x!)
        .attr('y2', (d) => (d.target as SimNode).y!);

      linkLabelGroup
        .attr('transform', (d) => {
          const sx = (d.source as SimNode).x!;
          const sy = (d.source as SimNode).y!;
          const tx = (d.target as SimNode).x!;
          const ty = (d.target as SimNode).y!;
          return `translate(${(sx + tx) / 2}, ${(sy + ty) / 2})`;
        });

      node.attr('transform', (d) => `translate(${d.x},${d.y})`);
    });
  }, [graphData]);

  if (loading) {
    return (
      <div className='flex items-center justify-center h-full text-stone-400'>
        グラフを構築中...
      </div>
    );
  }

  if (error) {
    return (
      <div className='flex flex-col items-center justify-center h-full gap-3 text-center px-4'>
        <p className='text-red-400 text-sm max-w-md'>{error}</p>
        <button
          onClick={fetchGraph}
          className='px-4 py-2 text-sm bg-stone-700 text-stone-200 rounded-lg hover:bg-stone-600 transition-colors'
        >
          再試行
        </button>
      </div>
    );
  }

  return (
    <div className='relative w-full h-full'>
      <svg ref={svgRef} className='w-full h-full bg-stone-900 rounded-lg' />

      {/* Legend */}
      <div className='absolute top-3 left-3 bg-stone-800/80 rounded-lg p-3 text-sm space-y-1.5'>
        {Object.entries(TYPE_COLORS).map(([type, color]) => (
          <div key={type} className='flex items-center gap-2'>
            <span
              className='inline-block w-3.5 h-3.5 rounded-full'
              style={{ backgroundColor: color }}
            />
            <span className='text-stone-300'>{type}</span>
          </div>
        ))}
      </div>

      {/* Selected entity panel */}
      {selected && (
        <div className='absolute bottom-3 left-3 bg-stone-800/90 rounded-lg p-4 max-w-xs'>
          <div className='flex items-center justify-between mb-2'>
            <h3 className='text-base font-bold text-stone-100'>{selected.name}</h3>
            <button
              onClick={() => setSelected(null)}
              className='text-stone-400 hover:text-stone-200 text-sm'
            >
              ✕
            </button>
          </div>
          <span
            className='inline-block text-sm px-2 py-0.5 rounded-full mb-2'
            style={{
              backgroundColor: nodeColor(selected.type) + '30',
              color: nodeColor(selected.type),
            }}
          >
            {selected.type}
          </span>
          <p className='text-sm text-stone-400 mb-2'>
            含まれる記事: {selected.article_ids.length}件
          </p>
          <div className='space-y-1'>
            {selected.article_ids.slice(0, 5).map((aid: string) => (
              <Link
                key={aid}
                href={`/wiki/${aid}`}
                className='block text-sm text-blue-400 hover:text-blue-300 truncate'
              >
                {aid}
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
