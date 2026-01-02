/**
 * stream_data 字段支持测试工具
 * 用于在浏览器控制台中测试 stream_data 字段的处理逻辑
 */

import { WuKongIMUtils } from '@/services/wukongimApi';
import type { WuKongIMMessage, WuKongIMMessagePayload } from '@/types';

/**
 * 创建测试用的 WuKongIM 消息
 */
const createTestMessage = (
  messageIdNum: number,
  content: string,
  streamData?: string | null
): WuKongIMMessage => {
  return {
    header: {
      no_persist: 0,
      red_dot: 1,
      sync_once: 0
    },
    setting: 0,
    message_id_str: String(messageIdNum),
    client_msg_no: `test-msg-${messageIdNum}`,
    message_seq: messageIdNum,
    from_uid: 'test-user',
    channel_id: 'test-channel',
    channel_type: 1,
    timestamp: Math.floor(Date.now() / 1000),
    payload: {
      content: content,
      type: 1
    },
    stream_data: streamData,
    end: streamData ? 0 : null,
    end_reason: null
  };
};

/**
 * 测试 stream_data 优先级
 */
export const testStreamDataPriority = (): void => {
  console.log('🧪 测试 stream_data 字段优先级...');
  console.log('');

  // 测试用例 1: stream_data 有值，应该优先使用
  const msg1 = createTestMessage(1, 'Payload content', 'Stream data content');
  const content1 = WuKongIMUtils.extractMessageContent(msg1);
  console.log('📋 测试用例 1: stream_data 有值');
  console.log('  payload.content:', 'Payload content');
  console.log('  stream_data:', 'Stream data content');
  console.log('  提取结果:', content1);
  console.log('  测试结果:', content1 === 'Stream data content' ? '✅ 通过' : '❌ 失败');
  console.log('');

  // 测试用例 2: stream_data 为空字符串，应该使用 payload
  const msg2 = createTestMessage(2, 'Payload content', '');
  const content2 = WuKongIMUtils.extractMessageContent(msg2);
  console.log('📋 测试用例 2: stream_data 为空字符串');
  console.log('  payload.content:', 'Payload content');
  console.log('  stream_data:', '""');
  console.log('  提取结果:', content2);
  console.log('  测试结果:', content2 === 'Payload content' ? '✅ 通过' : '❌ 失败');
  console.log('');

  // 测试用例 3: stream_data 为空白字符串，应该使用 payload
  const msg3 = createTestMessage(3, 'Payload content', '   ');
  const content3 = WuKongIMUtils.extractMessageContent(msg3);
  console.log('📋 测试用例 3: stream_data 为空白字符串');
  console.log('  payload.content:', 'Payload content');
  console.log('  stream_data:', '"   "');
  console.log('  提取结果:', content3);
  console.log('  测试结果:', content3 === 'Payload content' ? '✅ 通过' : '❌ 失败');
  console.log('');

  // 测试用例 4: stream_data 为 null，应该使用 payload
  const msg4 = createTestMessage(4, 'Payload content', null);
  const content4 = WuKongIMUtils.extractMessageContent(msg4);
  console.log('📋 测试用例 4: stream_data 为 null');
  console.log('  payload.content:', 'Payload content');
  console.log('  stream_data:', 'null');
  console.log('  提取结果:', content4);
  console.log('  测试结果:', content4 === 'Payload content' ? '✅ 通过' : '❌ 失败');
  console.log('');

  // 测试用例 5: stream_data 为 undefined（旧消息），应该使用 payload
  const msg5 = createTestMessage(5, 'Payload content', undefined);
  const content5 = WuKongIMUtils.extractMessageContent(msg5);
  console.log('📋 测试用例 5: stream_data 为 undefined（旧消息）');
  console.log('  payload.content:', 'Payload content');
  console.log('  stream_data:', 'undefined');
  console.log('  提取结果:', content5);
  console.log('  测试结果:', content5 === 'Payload content' ? '✅ 通过' : '❌ 失败');
  console.log('');

  console.log('✅ stream_data 优先级测试完成');
};

/**
 * 测试流式消息场景
 */
export const testStreamingScenario = (): void => {
  console.log('🧪 测试流式消息场景...');
  console.log('');

  // 模拟 AI 流式回复的多个消息片段
  const streamMessages = [
    createTestMessage(101, '', '你好'),
    createTestMessage(102, '', '你好，我是'),
    createTestMessage(103, '', '你好，我是 AI'),
    createTestMessage(104, '', '你好，我是 AI 助手'),
    createTestMessage(105, '', '你好，我是 AI 助手，有什么可以帮助你的吗？')
  ];

  console.log('📋 模拟 AI 流式回复场景:');
  streamMessages.forEach((msg, index) => {
    const content = WuKongIMUtils.extractMessageContent(msg);
    console.log(`  片段 ${index + 1}:`, content);
  });
  console.log('');

  // 最终消息（流式结束）
  const finalMessage: WuKongIMMessage = {
    ...streamMessages[4],
    end: 1,
    end_reason: null
  };
  const finalContent = WuKongIMUtils.extractMessageContent(finalMessage);
  console.log('📋 流式消息结束:');
  console.log('  最终内容:', finalContent);
  console.log('  end 标志:', finalMessage.end);
  console.log('  end_reason:', finalMessage.end_reason);
  console.log('');

  console.log('✅ 流式消息场景测试完成');
};

/**
 * 测试混合消息类型
 */
export const testMixedMessageTypes = (): void => {
  console.log('🧪 测试混合消息类型...');
  console.log('');

  const messages = [
    {
      name: '流式消息（AI 回复）',
      message: createTestMessage(201, '', 'AI 助手的流式回复内容'),
      expectedContent: 'AI 助手的流式回复内容'
    },
    {
      name: '普通文本消息',
      message: createTestMessage(202, '用户发送的普通消息', null),
      expectedContent: '用户发送的普通消息'
    },
    {
      name: '旧版本消息（无 stream_data）',
      message: {
        ...createTestMessage(203, '旧版本的消息', undefined),
        stream_data: undefined as any
      },
      expectedContent: '旧版本的消息'
    }
  ];

  messages.forEach(({ name, message, expectedContent }) => {
    const content = WuKongIMUtils.extractMessageContent(message);
    const passed = content === expectedContent;
    console.log(`📋 ${name}:`);
    console.log('  提取内容:', content);
    console.log('  预期内容:', expectedContent);
    console.log('  测试结果:', passed ? '✅ 通过' : '❌ 失败');
    console.log('');
  });

  console.log('✅ 混合消息类型测试完成');
};

/**
 * 测试边界情况
 */
export const testEdgeCases = (): void => {
  console.log('🧪 测试边界情况...');
  console.log('');

  const edgeCases = [
    {
      name: 'stream_data 和 payload 都为空',
      message: createTestMessage(301, '', ''),
      expectedContent: ''
    },
    {
      name: 'stream_data 包含特殊字符',
      message: createTestMessage(302, 'Normal', '特殊字符: \n\t"\'<>'),
      expectedContent: '特殊字符: \n\t"\'<>'
    },
    {
      name: 'stream_data 包含 emoji',
      message: createTestMessage(303, 'Normal', '你好 👋 世界 🌍'),
      expectedContent: '你好 👋 世界 🌍'
    },
    {
      name: 'stream_data 很长的文本',
      message: createTestMessage(304, 'Short', 'A'.repeat(1000)),
      expectedContent: 'A'.repeat(1000)
    },
    {
      name: 'stream_data 包含 HTML',
      message: createTestMessage(305, 'Normal', '<div>HTML content</div>'),
      expectedContent: '<div>HTML content</div>'
    }
  ];

  edgeCases.forEach(({ name, message, expectedContent }) => {
    const content = WuKongIMUtils.extractMessageContent(message);
    const passed = content === expectedContent;
    console.log(`📋 ${name}:`);
    console.log('  提取内容长度:', content.length);
    console.log('  预期内容长度:', expectedContent.length);
    console.log('  内容匹配:', passed ? '✅ 是' : '❌ 否');
    if (!passed && content.length < 100) {
      console.log('  实际内容:', content);
      console.log('  预期内容:', expectedContent);
    }
    console.log('');
  });

  console.log('✅ 边界情况测试完成');
};

/**
 * 测试消息转换
 */
export const testMessageConversion = (): void => {
  console.log('🧪 测试消息转换...');
  console.log('');

  const wkMessage = createTestMessage(
    401,
    'Payload content',
    'Stream data content (should be used)'
  );

  console.log('📋 原始 WuKongIM 消息:');
  console.log('  message_id_str:', wkMessage.message_id_str);
  console.log('  from_uid:', wkMessage.from_uid);
  console.log('  payload.content:', (wkMessage.payload as WuKongIMMessagePayload).content);
  console.log('  stream_data:', wkMessage.stream_data);
  console.log('');

  const convertedMessage = WuKongIMUtils.convertToMessage(wkMessage);

  console.log('📋 转换后的内部消息:');
  console.log('  id:', convertedMessage.id);
  console.log('  content:', convertedMessage.content);
  console.log('  sender type:', convertedMessage.type);
  console.log('  timestamp:', convertedMessage.timestamp);
  console.log('');

  const passed = convertedMessage.content === 'Stream data content (should be used)';
  console.log('测试结果:', passed ? '✅ 通过 - 正确使用了 stream_data' : '❌ 失败 - 未使用 stream_data');
  console.log('');

  console.log('✅ 消息转换测试完成');
};

/**
 * 运行所有测试
 */
export const runAllStreamDataTests = (): void => {
  console.log('🚀 开始运行 stream_data 字段支持测试...');
  console.log('');
  console.log('='.repeat(60));
  console.log('');

  testStreamDataPriority();
  console.log('='.repeat(60));
  console.log('');

  testStreamingScenario();
  console.log('='.repeat(60));
  console.log('');

  testMixedMessageTypes();
  console.log('='.repeat(60));
  console.log('');

  testEdgeCases();
  console.log('='.repeat(60));
  console.log('');

  testMessageConversion();
  console.log('='.repeat(60));
  console.log('');

  console.log('🎉 所有测试完成！');
  console.log('');
  console.log('💡 使用说明:');
  console.log('- testStreamDataPriority() - 测试 stream_data 优先级');
  console.log('- testStreamingScenario() - 测试流式消息场景');
  console.log('- testMixedMessageTypes() - 测试混合消息类型');
  console.log('- testEdgeCases() - 测试边界情况');
  console.log('- testMessageConversion() - 测试消息转换');
  console.log('- runAllStreamDataTests() - 运行所有测试');
};

// 在开发环境中将测试函数添加到全局对象
if (typeof window !== 'undefined' && import.meta.env.DEV) {
  (window as any).testStreamDataPriority = testStreamDataPriority;
  (window as any).testStreamingScenario = testStreamingScenario;
  (window as any).testMixedMessageTypes = testMixedMessageTypes;
  (window as any).testEdgeCases = testEdgeCases;
  (window as any).testMessageConversion = testMessageConversion;
  (window as any).runAllStreamDataTests = runAllStreamDataTests;
  
  console.log('🧪 stream_data 测试工具已加载');
  console.log('🧪 使用 runAllStreamDataTests() 运行所有测试');
}
