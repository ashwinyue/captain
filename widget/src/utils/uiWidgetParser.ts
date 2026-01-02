/**
 * TGO UI Widget 解析器
 * 用于解析 Markdown 内容中的 tgo-ui-widget 代码块
 */

import type { WidgetData } from '../components/widgets';

/**
 * 解析后的 UI Widget 块
 */
export interface ParsedUIBlock {
  type: string;
  data: WidgetData;
  raw: string;
  blockId: string;
  startIndex: number;
  endIndex: number;
}

/**
 * UI Widget 代码块正则表达式
 * 匹配 ```tgo-ui-widget 开头，``` 结尾的代码块
 */
const UI_WIDGET_REGEX = /```tgo-ui-widget\s*\n?([\s\S]*?)```/gi;

/**
 * 生成唯一 ID
 */
function generateBlockId(): string {
  return `ui-block-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * 解析 Markdown 内容中的所有 UI Widget 块
 */
export function parseUIWidgets(content: string): ParsedUIBlock[] {
  const blocks: ParsedUIBlock[] = [];
  let match: RegExpExecArray | null;

  // Reset regex lastIndex
  UI_WIDGET_REGEX.lastIndex = 0;

  while ((match = UI_WIDGET_REGEX.exec(content)) !== null) {
    const rawJson = match[1].trim();

    try {
      const data = JSON.parse(rawJson) as WidgetData;

      if (!data.type) {
        console.warn('UI Widget missing type field:', rawJson);
        continue;
      }

      blocks.push({
        type: data.type,
        data,
        raw: rawJson,
        blockId: generateBlockId(),
        startIndex: match.index,
        endIndex: match.index + match[0].length,
      });
    } catch (e) {
      console.error('Failed to parse UI Widget JSON:', e);
    }
  }

  return blocks;
}

/**
 * 替换 Markdown 内容中的 UI Widget 块为占位符
 * 返回处理后的内容和块映射
 */
export function replaceUIWidgetsWithPlaceholders(
  content: string
): {
  content: string;
  blocks: Map<string, ParsedUIBlock>;
} {
  const blocks = new Map<string, ParsedUIBlock>();

  // Reset regex lastIndex
  UI_WIDGET_REGEX.lastIndex = 0;

  const processedContent = content.replace(
    UI_WIDGET_REGEX,
    (match, jsonContent) => {
      try {
        // Trim the JSON content and remove any trailing newlines before the closing ```
        const trimmedJson = jsonContent.trim();
        const data = JSON.parse(trimmedJson) as WidgetData;
        const blockId = generateBlockId();

        blocks.set(blockId, {
          type: data.type,
          data,
          raw: trimmedJson,
          blockId,
          startIndex: 0,
          endIndex: 0,
        });

        // Return a placeholder that can be replaced with the rendered component
        return `\n\n<div data-ui-widget="${blockId}"></div>\n\n`;
      } catch (e) {
        // If parsing fails, log and return original content
        console.error('Failed to parse UI Widget JSON:', e, 'Content:', jsonContent?.substring(0, 100));
        return match;
      }
    }
  );

  return { content: processedContent, blocks };
}

/**
 * 检查内容是否包含 UI Widget
 */
export function hasUIWidgets(content: string): boolean {
  UI_WIDGET_REGEX.lastIndex = 0;
  return UI_WIDGET_REGEX.test(content);
}

/**
 * 流式解析器类 - 用于处理流式内容
 */
export class StreamingUIWidgetParser {
  private buffer = '';
  private inBlock = false;
  private blockContent = '';
  private blockStartIndex = 0;

  /**
   * 喂入一个 chunk
   * @returns 解析出的完整块（如果有）
   */
  feed(chunk: string): {
    nonWidgetContent: string;
    completedBlocks: ParsedUIBlock[];
  } {
    this.buffer += chunk;
    const completedBlocks: ParsedUIBlock[] = [];
    let nonWidgetContent = '';

    while (this.buffer.length > 0) {
      if (!this.inBlock) {
        // 查找块开始
        const startMatch = this.buffer.match(/```tgo-ui-widget\s*\n?/i);

        if (startMatch && startMatch.index !== undefined) {
          // 输出块开始前的内容
          nonWidgetContent += this.buffer.slice(0, startMatch.index);
          this.buffer = this.buffer.slice(startMatch.index + startMatch[0].length);
          this.inBlock = true;
          this.blockContent = '';
          this.blockStartIndex = startMatch.index;
        } else {
          // 检查是否可能是部分匹配
          const potentialStart = this.buffer.lastIndexOf('```');
          if (potentialStart >= 0 && potentialStart > this.buffer.length - 20) {
            nonWidgetContent += this.buffer.slice(0, potentialStart);
            this.buffer = this.buffer.slice(potentialStart);
          } else {
            nonWidgetContent += this.buffer;
            this.buffer = '';
          }
          break;
        }
      } else {
        // 在块内，查找结束
        const endIndex = this.buffer.indexOf('\n```');

        if (endIndex >= 0) {
          // 找到结束
          this.blockContent += this.buffer.slice(0, endIndex);
          this.buffer = this.buffer.slice(endIndex + 4); // Skip \n```
          this.inBlock = false;

          // 解析块
          try {
            const data = JSON.parse(this.blockContent.trim()) as WidgetData;
            completedBlocks.push({
              type: data.type,
              data,
              raw: this.blockContent.trim(),
              blockId: generateBlockId(),
              startIndex: this.blockStartIndex,
              endIndex: this.blockStartIndex + this.blockContent.length,
            });
          } catch (e) {
            console.error('Failed to parse streaming UI Widget:', e);
          }
        } else {
          // 还没找到结束，保持缓冲
          // 但要检查是否可能是部分结束标记
          if (this.buffer.length > 3) {
            this.blockContent += this.buffer.slice(0, -3);
            this.buffer = this.buffer.slice(-3);
          }
          break;
        }
      }
    }

    return { nonWidgetContent, completedBlocks };
  }

  /**
   * 刷新缓冲区
   */
  flush(): string {
    const remaining = this.buffer;
    this.buffer = '';
    this.inBlock = false;
    this.blockContent = '';
    return remaining;
  }

  /**
   * 是否在块内
   */
  get isInBlock(): boolean {
    return this.inBlock;
  }
}
