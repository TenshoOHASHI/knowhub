import type { Plugin } from 'unified';
import type { Html, Paragraph, Root } from 'mdast';

export type CalloutType = 'note' | 'info' | 'tip' | 'warning' | 'caution' | 'important' | 'warm';

const GITHUB_TYPE_MAP: Record<string, CalloutType> = {
  NOTE: 'note',
  INFO: 'info',
  TIP: 'tip',
  WARNING: 'warning',
  CAUTION: 'caution',
  IMPORTANT: 'important',
};

const ZENN_TYPE_MAP: Record<string, CalloutType> = {
  '': 'note',
  alert: 'caution',
  info: 'info',
  tip: 'tip',
  warning: 'warning',
  important: 'important',
  warm: 'warm',
};

const CALLOUT_TYPES = Object.keys(GITHUB_TYPE_MAP).join('|');
const CALLOUT_RE = new RegExp(`^\\[!(${CALLOUT_TYPES})\\]\\s*\\n?`);
const ZENN_TYPES = Object.keys(ZENN_TYPE_MAP).filter(Boolean).join('|');
const ZENN_RE = new RegExp(`^:::message\\s*(?:(${ZENN_TYPES}))?\\s*\\n([\\s\\S]*?)^:::\\s*$`, 'gm');

/**
 * Pre-process raw markdown to convert Zenn-style callouts to HTML divs.
 * :::message ... ::: → <div class="callout callout-note">...</div>
 * :::message alert ... ::: → <div class="callout callout-caution">...</div>
 */
export function preprocessCallouts(markdown: string): string {
  return markdown.replace(
    ZENN_RE,
    (_match, typeFlag, content) => {
      const type = ZENN_TYPE_MAP[typeFlag || ''] || 'note';
      return `<div class="callout callout-${type}">\n\n${content.trim()}\n\n</div>`;
    }
  );
}

/**
 * Remark plugin that transforms GitHub-style blockquote callouts (> [!TYPE])
 * into HTML div wrappers so the Callout component can render them.
 */
export const remarkCallout: Plugin<[], Root> = function () {
  return (tree: Root) => {
    const newChildren: typeof tree.children = [];

    for (const node of tree.children) {
      if (node.type === 'blockquote') {
        const children = node.children;
        if (children.length > 0) {
          const first = children[0];
          if (first.type === 'paragraph' && first.children.length > 0) {
            const firstChild = first.children[0];
            if (firstChild.type === 'text') {
              const match = (firstChild.value as string).match(CALLOUT_RE);
              if (match) {
                const calloutType = GITHUB_TYPE_MAP[match[1]];
                if (calloutType) {
                  const restText = (firstChild.value as string).slice(match[0].length);
                  const contentNodes: typeof children = [];

                  if (restText.trim()) {
                    const paragraph: Paragraph = {
                      type: 'paragraph',
                      children: [{ type: 'text', value: restText.replace(/^\n/, '') }],
                    };
                    contentNodes.push(paragraph);
                  }

                  contentNodes.push(...children.slice(1));

                  const openCallout: Html = {
                    type: 'html',
                    value: `<div class="callout callout-${calloutType}">`,
                  };
                  newChildren.push(openCallout);
                  newChildren.push(...contentNodes);
                  const closeCallout: Html = {
                    type: 'html',
                    value: '</div>',
                  };
                  newChildren.push(closeCallout);
                  continue;
                }
              }
            }
          }
        }
      }

      newChildren.push(node);
    }

    tree.children = newChildren;
  };
};
