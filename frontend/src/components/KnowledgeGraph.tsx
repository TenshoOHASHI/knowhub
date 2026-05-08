'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import * as d3 from 'd3';
import {
  getKnowledgeGraph,
  getArticles,
  getRelatedArticles,
  type EntityNode,
  type RelationEdge,
  type RelatedArticle,
} from '@/lib/api';
import Link from 'next/link';

interface SimNode extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: string;
  articleIds: string[];
  connections: number; // 接続数
}

interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  label: string;
}

interface GraphStats {
  nodeCount: number;
  edgeCount: number;
  maxConnections: number;
  topHub: { name: string; connections: number };
  typeDistribution: Record<string, number>;
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
  const [highlighted, setHighlighted] = useState<Set<string>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<EntityNode[]>([]);
  const [showSearch, setShowSearch] = useState(false);
  const [articles, setArticles] = useState<Map<string, { title: string; id: string }>>(new Map());
  const [relatedArticles, setRelatedArticles] = useState<RelatedArticle[]>([]);
  const [showStats, setShowStats] = useState(true);
  const [graphStats, setGraphStats] = useState<GraphStats | null>(null);

  const [graphData, setGraphData] = useState<{
    entities: EntityNode[];
    relations: RelationEdge[];
  } | null>(null);

  // 記事マップを構築
  useEffect(() => {
    getArticles()
      .then((data: { articles: { id: string; title: string }[] }) => {
        const map = new Map();
        data.articles?.forEach((a) => map.set(a.id, { title: a.title, id: a.id }));
        setArticles(map);
      })
      .catch(() => {
        // 記事取得エラーは無視
      });
  }, []);

  // 関連記事を取得
  useEffect(() => {
    if (!selected) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setRelatedArticles([]);
      return;
    }

    // 選択エンティティが含まれる最初の記事を取得
    const firstArticleId = selected.article_ids?.[0];
    if (!firstArticleId) return;

    getRelatedArticles(firstArticleId, 2, 5)
      .then((data) => {
        setRelatedArticles(data.results || []);
      })
      .catch(() => {
        setRelatedArticles([]);
      });
  }, [selected]);

  // 統計計算（fetchGraphの前に宣言）
  const calculateStats = useCallback((entities: EntityNode[], relations: RelationEdge[]): GraphStats => {
    const nodeCount = entities.length;
    const edgeCount = relations.length;

    // 接続数を計算
    const connectionMap = new Map<string, number>();
    relations.forEach((r) => {
      connectionMap.set(r.source, (connectionMap.get(r.source) || 0) + 1);
      connectionMap.set(r.target, (connectionMap.get(r.target) || 0) + 1);
    });

    let maxConnections = 0;
    let topHub = { name: '', connections: 0 };
    connectionMap.forEach((count, id) => {
      if (count > maxConnections) {
        maxConnections = count;
        const entity = entities.find((e) => e.id === id);
        if (entity) {
          topHub = { name: entity.name, connections: count };
        }
      }
    });

    // タイプ別分布
    const typeDistribution: Record<string, number> = {};
    entities.forEach((e) => {
      typeDistribution[e.type] = (typeDistribution[e.type] || 0) + 1;
    });

    return {
      nodeCount,
      edgeCount,
      maxConnections,
      topHub,
      typeDistribution,
    };
  }, []);

  const fetchGraph = useCallback(() => {
    setLoading(true);
    setError(null);
    setGraphData(null);

    let cancelled = false;

    getKnowledgeGraph()
      .then((data) => {
        if (cancelled) return;
        setLoading(false);
        setGraphData(data);

        // グラフ統計を計算
        const stats = calculateStats(data.entities, data.relations);
        setGraphStats(stats);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err.message);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [calculateStats]);

  // 検索
  useEffect(() => {
    if (!searchQuery || !graphData) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSearchResults([]);
      return;
    }

    const query = searchQuery.toLowerCase();
    const results = graphData.entities.filter(
      (e) => e.name.toLowerCase().includes(query) || e.type.toLowerCase().includes(query)
    );
    setSearchResults(results);
  }, [searchQuery, graphData]);

  // 検索結果を選択
  const selectSearchResult = (entity: EntityNode) => {
    setSearchQuery('');
    setShowSearch(false);
    setSelected(entity);
    // グラフを再描画してハイライト
  };

  // 初回フェッチ
  useEffect(() => {
    let cancelled = false;

    getKnowledgeGraph()
      .then((data) => {
        if (cancelled) return;
        setLoading(false);
        setGraphData(data);

        const stats = calculateStats(data.entities, data.relations);
        setGraphStats(stats);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err.message);
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [calculateStats]);

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

    // 接続数を計算
    const connectionMap = new Map<string, number>();
    relations.forEach((r) => {
      connectionMap.set(r.source, (connectionMap.get(r.source) || 0) + 1);
      connectionMap.set(r.target, (connectionMap.get(r.target) || 0) + 1);
    });

    const nodes: SimNode[] = entities.map((e) => ({
      id: e.id,
      name: e.name,
      type: e.type,
      articleIds: e.article_ids ?? [],
      connections: connectionMap.get(e.id) || 0,
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

    // 隣接ノードを取得
    const getNeighbors = (nodeId: string): Set<string> => {
      const neighbors = new Set<string>();
      links.forEach((l) => {
        if (l.source === nodeId) neighbors.add(l.target as string);
        if (l.target === nodeId) neighbors.add(l.source as string);
      });
      return neighbors;
    };

    const simulation = d3
      .forceSimulation<SimNode>(nodes)
      .force(
        'link',
        d3
          .forceLink<SimNode, SimLink>(links)
          .id((d) => d.id)
          .distance(120),
      )
      .force('charge', d3.forceManyBody().strength(-400))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('x', d3.forceX(width / 2).strength(0.05))
      .force('y', d3.forceY(height / 2).strength(0.05))
      .force(
        'collision',
        d3.forceCollide<SimNode>().radius((d) => nodeRadius(d) + 8),
      )
      .alphaDecay(0.02);

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

    // ドロップシャドウ
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

    // Circles
    node
      .append('circle')
      .attr('r', nodeRadius)
      .attr('fill', (d) => nodeColor(d.type))
      .attr('stroke', '#fff')
      .attr('stroke-width', 2.5)
      .attr('filter', 'url(#shadow)')
      .attr('class', 'node-circle')
      .style('cursor', 'pointer')
      .on('click', (event, d) => {
        event.stopPropagation();
        const neighbors = getNeighbors(d.id);
        setHighlighted(neighbors);
        setSelected({
          id: d.id,
          name: d.name,
          type: d.type,
          article_ids: d.articleIds,
        });
      })
      .on('dblclick', (event, _d) => {
        event.stopPropagation();
        // ダブルクリックでハイライト解除
        setHighlighted(new Set());
      });

    // Labels
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

    // 接続数バッジ
    node
      .append('text')
      .text((d) => d.connections.toString())
      .attr('font-size', 10)
      .attr('font-weight', 700)
      .attr('text-anchor', 'middle')
      .attr('dy', 4)
      .attr('fill', '#fff')
      .style('pointer-events', 'none')
      .style('opacity', 0)
      .attr('class', 'connection-badge');

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

    // ハイライト更新
    const updateHighlight = () => {
      const isHighlighted = (d: SimNode) =>
        highlighted.size === 0 || highlighted.has(d.id);

      node.selectAll('.node-circle').transition().duration(200)
        .attr('fill', (d) => isHighlighted(d) ? nodeColor(d.type) : '#374151')
        .attr('opacity', (d) => isHighlighted(d) ? 1 : 0.3)
        .attr('r', (d) => isHighlighted(d) ? nodeRadius(d) : nodeRadius(d) * 0.8);

      node.selectAll('text').transition().duration(200)
        .attr('opacity', (d) => isHighlighted(d) ? 1 : 0.2);

      link.transition().duration(200)
        .attr('stroke-opacity', (d) => {
          const sourceId = (d.source as SimNode).id;
          const targetId = (d.target as SimNode).id;
          if (highlighted.size === 0) return 0.4;
          return highlighted.has(sourceId) && highlighted.has(targetId) ? 0.8 : 0.1;
        })
        .attr('stroke', (d) => {
          const sourceId = (d.source as SimNode).id;
          const targetId = (d.target as SimNode).id;
          if (highlighted.size === 0) return '#64748b';
          return highlighted.has(sourceId) && highlighted.has(targetId) ? '#fbbf24' : '#64748b';
        });

      linkLabelGroup.transition().duration(200)
        .attr('opacity', (d) => {
          const sourceId = (d.source as SimNode).id;
          const targetId = (d.target as SimNode).id;
          if (highlighted.size === 0) return 1;
          return highlighted.has(sourceId) && highlighted.has(targetId) ? 1 : 0.1;
        });
    };

    // SVGクリックでハイライト解除
    svg.on('click', () => {
      setHighlighted(new Set());
      setSelected(null);
    });

    // ハイライト更新を定期的にチェック（簡易実装）
    const highlightInterval = setInterval(updateHighlight, 100);

    // シミュレーション終了後に全体が中央に収まるようにズーム
    simulation.on('end', () => {
      clearInterval(highlightInterval);
      updateHighlight();

      const padding = 60;
      const bbox = (g.node() as SVGGElement).getBBox();
      if (bbox.width === 0 || bbox.height === 0) return;

      const scale = Math.min(
        (width - padding * 2) / bbox.width,
        (height - padding * 2) / bbox.height,
        1.5,
      );
      const tx = width / 2 - (bbox.x + bbox.width / 2) * scale;
      const ty = height / 2 - (bbox.y + bbox.height / 2) * scale;

      svg.transition().duration(500).call(
        zoom.transform,
        d3.zoomIdentity.translate(tx, ty).scale(scale),
      );
    });

    return () => {
      clearInterval(highlightInterval);
    };
  }, [graphData, highlighted]);

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

      {/* 統計パネル */}
      {showStats && graphStats && (
        <div className='absolute top-3 left-3 bg-stone-800/90 rounded-lg p-4 text-sm min-w-[220px] max-w-[280px]'>
          <div className='flex items-center justify-between mb-3'>
            <h3 className='text-base font-bold text-stone-100'>グラフ統計</h3>
            <button
              onClick={() => setShowStats(false)}
              className='text-stone-400 hover:text-stone-200 text-xs'
            >
              ✕
            </button>
          </div>

          <div className='space-y-2 text-stone-300'>
            <div className='flex justify-between'>
              <span>ノード数:</span>
              <span className='font-mono text-blue-400'>{graphStats.nodeCount}</span>
            </div>
            <div className='flex justify-between'>
              <span>エッジ数:</span>
              <span className='font-mono text-purple-400'>{graphStats.edgeCount}</span>
            </div>
            <div className='flex justify-between'>
              <span>最大接続数:</span>
              <span className='font-mono text-green-400'>{graphStats.maxConnections}</span>
            </div>
            <div className='pt-2 border-t border-stone-600'>
              <div className='text-xs text-stone-400 mb-1'>最多接続ハブ</div>
              <div className='text-amber-400 font-semibold truncate'>
                {graphStats.topHub.name}
              </div>
              <div className='text-xs text-stone-500'>
                {graphStats.topHub.connections} 接続
              </div>
            </div>

            {/* タイプ別分布 */}
            <div className='pt-2 border-t border-stone-600'>
              <div className='text-xs text-stone-400 mb-2'>タイプ別分布</div>
              <div className='space-y-1'>
                {Object.entries(graphStats.typeDistribution)
                  .sort(([, a], [, b]) => b - a)
                  .slice(0, 5)
                  .map(([type, count]) => (
                    <div key={type} className='flex items-center gap-2'>
                      <span
                        className='inline-block w-2.5 h-2.5 rounded-full'
                        style={{ backgroundColor: nodeColor(type) }}
                      />
                      <span className='flex-1 text-xs truncate'>{type}</span>
                      <span className='font-mono text-xs'>{count}</span>
                    </div>
                  ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 統計パネル閉じている時の再表示ボタン */}
      {!showStats && (
        <button
          onClick={() => setShowStats(true)}
          className='absolute top-3 left-3 bg-stone-800/90 rounded-lg p-2 text-stone-300 hover:text-stone-100 hover:bg-stone-700 transition-colors'
          title='統計を表示'
        >
          <svg className='w-5 h-5' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
            <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' />
          </svg>
        </button>
      )}

      {/* 検索ボタン */}
      <button
        onClick={() => setShowSearch(!showSearch)}
        className='absolute top-3 right-3 bg-stone-800/90 rounded-lg p-2 text-stone-300 hover:text-stone-100 hover:bg-stone-700 transition-colors'
        title='検索'
      >
        <svg className='w-5 h-5' fill='none' stroke='currentColor' viewBox='0 0 24 24'>
          <path strokeLinecap='round' strokeLinejoin='round' strokeWidth={2} d='M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z' />
        </svg>
      </button>

      {/* 検索パネル */}
      {showSearch && (
        <div className='absolute top-14 right-3 bg-stone-800/90 rounded-lg p-3 w-64'>
          <input
            type='text'
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder='エンティティ名を検索...'
            className='w-full px-3 py-2 bg-stone-700 text-stone-100 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500'
            autoFocus
          />
          {searchResults.length > 0 && (
            <div className='mt-2 max-h-48 overflow-y-auto'>
              {searchResults.map((entity) => (
                <button
                  key={entity.id}
                  onClick={() => selectSearchResult(entity)}
                  className='w-full text-left px-3 py-2 text-sm text-stone-300 hover:bg-stone-700 rounded-lg truncate'
                >
                  <span className='text-stone-400'>[{entity.type}]</span> {entity.name}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Legend */}
      <div className='absolute bottom-3 right-3 bg-stone-800/80 rounded-lg p-3 text-sm'>
        <div className='text-xs text-stone-400 mb-2'>ノードタイプ</div>
        <div className='grid grid-cols-2 gap-x-3 gap-y-1'>
          {Object.entries(TYPE_COLORS).map(([type, color]) => (
            <div key={type} className='flex items-center gap-2'>
              <span
                className='inline-block w-3 h-3 rounded-full'
                style={{ backgroundColor: color }}
              />
              <span className='text-stone-300 text-xs'>{type}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Selected entity panel */}
      {selected && (
        <div className='absolute bottom-3 left-3 bg-stone-800/95 rounded-lg p-4 max-w-sm max-h-[60vh] overflow-y-auto'>
          <div className='flex items-center justify-between mb-2'>
            <h3 className='text-base font-bold text-stone-100'>{selected.name}</h3>
            <button
              onClick={() => {
                setSelected(null);
                setHighlighted(new Set());
              }}
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
          <p className='text-sm text-stone-400 mb-3'>
            含まれる記事: {selected.article_ids.length}件
          </p>

          {/* 記事リスト（タイトル表示） */}
          <div className='space-y-2 mb-4'>
            <div className='text-xs text-stone-500 uppercase tracking-wide'>含まれる記事</div>
            {selected.article_ids.slice(0, 5).map((aid: string) => {
              const article = articles.get(aid);
              return (
                <Link
                  key={aid}
                  href={`/wiki/${aid}`}
                  className='block text-sm text-blue-400 hover:text-blue-300'
                >
                  <div className='truncate'>{article?.title || aid}</div>
                  {!article && <div className='text-xs text-stone-500'>{aid}</div>}
                </Link>
              );
            })}
            {selected.article_ids.length > 5 && (
              <div className='text-xs text-stone-500'>
                他 {selected.article_ids.length - 5} 件
              </div>
            )}
          </div>

          {/* 関連記事パス */}
          {relatedArticles.length > 0 && (
            <div className='pt-3 border-t border-stone-600'>
              <div className='text-xs text-stone-500 uppercase tracking-wide mb-2'>
                関連記事パス
              </div>
              <div className='space-y-2'>
                {relatedArticles.slice(0, 3).map((article) => (
                  <div key={article.article_id} className='bg-stone-700/50 rounded p-2'>
                    <Link
                      href={`/wiki/${article.article_id}`}
                      className='block text-sm text-blue-400 hover:text-blue-300 truncate mb-1'
                    >
                      {article.title}
                    </Link>
                    <div className='flex items-center gap-2'>
                      <div className='flex-1 h-1 bg-stone-600 rounded-full overflow-hidden'>
                        <div
                          className='h-full bg-green-500'
                          style={{ width: `${Math.min(article.relevance_score * 10, 100)}%` }}
                        />
                      </div>
                      <span className='text-xs text-stone-400'>
                        {article.relevance_score.toFixed(1)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
