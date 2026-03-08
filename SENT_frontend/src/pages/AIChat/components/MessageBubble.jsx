import ThinkingBlock from './ThinkingBlock';

const MessageBubble = ({ message }) => {
    const isAI = message.sender === 'ai';

    return (
        <div className={`flex flex-col mb-6 w-full ${isAI ? 'items-start' : 'items-end'}`}>
            {/* Tăng max-w lên 85% và thêm w-full để kiểm soát layout */}
            <div className={`flex gap-4 w-full max-w-[85%] ${isAI ? 'flex-row' : 'flex-row-reverse'}`}>
                
                {/* Thêm min-w-0 để ngăn chặn Flexbox tự động bung rộng quá màn hình khi có text dài */}
                <div className="flex flex-col min-w-0 w-full">
                    {/* HIỂN THỊ SUY NGHĨ */}
                    {isAI && message.thought && <ThinkingBlock thought={message.thought} />}

                    {/* NỘI DUNG CHÍNH (Đã thêm break-words và whitespace-pre-wrap) */}
                    <div 
                        className={`px-5 py-3.5 rounded-2xl text-sm leading-relaxed shadow-md whitespace-pre-wrap break-words ${
                            isAI 
                            ? 'bg-[#1e293b] text-slate-200 border border-slate-700/50' 
                            : 'bg-indigo-600 text-white'
                        }`}
                        style={{ wordBreak: 'break-word' }}
                    >
                        {message.text}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default MessageBubble;