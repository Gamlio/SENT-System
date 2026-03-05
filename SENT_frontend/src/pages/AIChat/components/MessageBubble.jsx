import ThinkingBlock from './ThinkingBlock';

const MessageBubble = ({ message }) => {
    const isAI = message.sender === 'ai';

    return (
        <div className={`flex flex-col mb-6 ${isAI ? 'items-start' : 'items-end'}`}>
            <div className={`flex gap-4 max-w-[80%] ${isAI ? 'flex-row' : 'flex-row-reverse'}`}>
                {/* Avatar... */}
                
                <div className="flex flex-col">
                    {/* HIỂN THỊ SUY NGHĨ (Chỉ AI mới có) */}
                    {isAI && message.thought && <ThinkingBlock thought={message.thought} />}

                    {/* NỘI DUNG CHÍNH */}
                    <div className={`px-5 py-3.5 rounded-2xl text-sm leading-relaxed shadow-md ${
                        isAI ? 'bg-[#1e293b] text-slate-200' : 'bg-emerald-600 text-white'
                    }`}>
                        {message.text}
                    </div>
                </div>
            </div>
        </div>
    );
};
export default MessageBubble;