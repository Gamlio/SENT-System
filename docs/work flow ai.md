Framework 6 lớp Claude Code : Luật – Bộ nhớ – Kỹ năng – Tác tử – Kiểm chứng – Tiến hóa
Làm sao để AI không chỉ trả lời hay, mà làm việc ổn định, nhớ việc cũ, phối hợp đúng vai và tạo ra kết quả đáng tin cho công việc cá nhân, ứng dụng trong doanh nghiệp?
1. Lớp Luật: phải định nghĩa “cách làm việc”, không chỉ “nhiệm vụ”
Bắt đầu từ CLAUDE. md và các rules luôn được nạp vào context, đặt ra delivery standards rất cụ thể: truth hơn speed, tự verify trước khi tuyên bố hoàn thành, không dùng các câu như “should be fine”, “probably passes”, và phải có rollback path. Tức là asset không được phép chỉ “có vẻ đúng”, mà phải làm việc như một người có kỷ luật vận hành. 
Đây là bài học đầu tiên khi xây skills/asset cho cá nhân hoặc doanh nghiệp: muốn asset tốt, phải dạy nó cách làm việc trước khi dạy nó làm việc gì?
Rất nhiều đội ngũ hiện nay xây asset theo kiểu: có một prompt vai trò, thêm vài knowledge file, xong. Cách đó thường tạo ra một “AI biết nhiều”, nhưng không tạo ra một “AI làm việc đáng tin”. Bởi điều thiếu không phải là tri thức, mà là luật hành vi.
Với cá nhân, lớp luật nên trả lời các câu hỏi như:
AI của tôi ưu tiên tốc độ hay độ chính xác?
Khi nào được phép kết luận?
Khi nào phải hỏi lại?
Khi nào phải ghi lại tri thức mới?
Với doanh nghiệp, lớp luật cần tiến thêm một bước:
khi nào asset được truy cập dữ liệu nào,
khi nào được đề xuất nhưng không được tự động hành động,
khi nào bắt buộc phải có human review,
và tiêu chuẩn nào để kết quả được xem là “đủ tốt để đưa vào vận hành”.
asset không thể trưởng thành nếu tổ chức chưa định nghĩa được kỷ luật làm việc cho asset.
2. Lớp Bộ nhớ: đừng để mọi phiên làm việc bắt đầu lại từ số 0
Tổ chức bộ nhớ thành nhiều tầng: today.md, projects.md, goals.md, active-tasks.json, cùng nguyên tắc SSOT – mỗi loại thông tin chỉ có một nơi chuẩn để ghi và cập nhật. Repo còn nêu rõ loại thông tin nào phải sống ở file nào, để tránh trùng lặp và lỗi thời.  Từ đó, insight rút ra là: năng lực dài hạn của asset không nằm ở model, mà nằm ở kiến trúc bộ nhớ.
Khi xây cho cá nhân, bộ nhớ nên có ít nhất 4 lớp:
. Một là bộ nhớ ngắn hạn: hôm nay đang làm gì, đang vướng gì, bước tiếp theo là gì.
. Hai là bộ nhớ dự án: mục tiêu, trạng thái, quyết định quan trọng, lỗi hay gặp.
. Ba là bộ nhớ mẫu hình: những pattern lặp lại có thể tái sử dụng.
. Bốn là bộ nhớ nhiệm vụ đang dang dở: việc nào đang active, blocked, waiting.
Khi xây cho doanh nghiệp, lớp bộ nhớ phải trở thành bản đồ tri thức vận hành. Lúc này không còn là today. md hay projects. md đơn giản nữa mà là:
• SOP và workflow của từng phòng ban
• policy và rule của tổ chức
• dữ liệu chuẩn vận hành
• incident log / lesson learned
• backlog cải tiến
• handoff giữa người và asset
Nếu không có bộ nhớ như vậy, doanh nghiệp chỉ đang dùng AI như một công cụ hỏi đáp. Nhưng khi có bộ nhớ đúng, AI bắt đầu chuyển thành một thành viên có lịch sử làm việc.
Và đây là điểm cực kỳ quan trọng: không nên lưu mọi thứ vào một đống knowledge base chung.
SSOT chính là để chống hiện tượng “một sự thật, nhiều bản sao”. Với doanh nghiệp, đây là lý do phải tách rõ: dữ liệu chuẩn ở đâu, policy ở đâu, bài học vận hành ở đâu, task đang chạy ở đâu. 
3. Lớp Kỹ năng: mỗi skill chỉ nên giải một loại công việc thật rõ
skills/ không phải nơi chứa các mẹo vặt. Nó chứa những năng lực có thể tái sử dụng như verification-before-completion, systematic-debugging, planning-with-files, experience-evolution, session-end. Tức là skill ở đây không phải “thông tin”, mà là một đơn vị năng lực hành động có cấu trúc. 
Từ đó, nguyên tắc xây skill nên là: Một skill = một job to be done rõ ràng.
Đừng làm skill kiểu:
“hỗ trợ marketing toàn diện”,
“giúp tôi phân tích dữ liệu”,
“viết nội dung và lập kế hoạch”.
Đó không phải skill. Đó là một wishlist.
Một skill tốt phải có 5 thành phần:
Mục tiêu rõ: skill này sinh ra để làm gì.
Trigger rõ: khi nào nên gọi nó.
Input rõ: cần những loại dữ liệu nào.
Quy trình rõ: các bước xử lý tuần tự.
Output rõ: kết quả trả về ở định dạng nào.
Nếu áp dụng cho cá nhân, có thể chia skill thành 5 nhóm:
Skill nghiên cứu: tổng hợp nguồn, trích insight, đối chiếu thông tin.
Skill sản xuất: viết proposal, soạn outline, tạo nội dung, chuẩn hóa đầu ra.
Skill kiểm tra: QA, fact check, review logic, review cấu trúc.
Skill điều phối: lập kế hoạch, chia nhỏ task, tạo handoff.
Skill học hỏi: ghi lại lesson learned, cập nhật memory, gom pattern.
Nếu áp dụng cho doanh nghiệp, nên chia theo năng lực vận hành:
Skill xử lý SOP
Skill phân tích dữ liệu vận hành
Skill hỗ trợ ra quyết định
Skill kiểm tra tuân thủ
Skill báo cáo và handoff
Skill quản trị tri thức
Điều đáng chú ý là: doanh nghiệp không nên bắt đầu bằng “asset lớn”, mà nên bắt đầu bằng “skill nhỏ nhưng đáng tin”.
Bởi asset lớn thường ấn tượng trong demo nhưng khó kiểm soát trong thực tế. Skill nhỏ thì dễ test, dễ đo, dễ cải tiến, và dễ ghép thành hệ thống sau này.
4. Lớp Tác tử: asset không phải “AI biết tuốt”, mà là vai trò chuyên môn hóa
assets/ với các asset chuyên vai như pr-reviewer, security-reviewer, performance-analyzer. Ý nghĩa của nó rất rõ: thay vì kỳ vọng một AI làm mọi thứ, tác giả tách theo chuyên môn và góc nhìn kiểm tra. 
Đây là điểm nhiều đội ngũ bỏ qua khi xây asset doanh nghiệp. Họ muốn một “siêu trợ lý” biết mọi thứ từ kế toán đến sales đến pháp lý. Kết quả là asset trả lời cái gì cũng được, nhưng cái gì cũng nông.
Cách đúng hơn là thiết kế asset như các vai trong một tổ chức.
Với cá nhân, anh có thể có các asset như:
• asset nghiên cứu thị trường
• asset kiến trúc giải pháp
• asset viết proposal
• asset review logic
• asset quản lý tri thức
• asset chuẩn bị họp / follow-up
Với doanh nghiệp, mô hình tác tử nên gần với cấu trúc phòng ban hoặc chức năng:
• asset tiếp nhận yêu cầu
• asset phân loại và định tuyến
• asset thực thi theo SOP
• asset kiểm tra tuân thủ
• asset tổng hợp báo cáo
• asset escalation sang con người
Quan trọng nhất: asset không nên được định nghĩa theo tên hay phong cách, mà theo trách nhiệm và ranh giới quyết định.
Một asset tốt cần biết 4 điều:
. Tôi chịu trách nhiệm phần nào.
. Tôi không được làm phần nào.
. Khi nào tôi phải gọi skill nào.
. Khi nào tôi phải chuyển tiếp cho asset khác hoặc con người.
Khi thiết kế như vậy, hệ asset bắt đầu giống một tổ chức thu nhỏ. Và lúc đó AI không còn là một chatbot, mà trở thành một hệ vai trò có điều phối.
5. Lớp Kiểm chứng: đây là cổng biến AI từ “thuyết phục” thành “đáng tin”
Skill verification-before-completion của repo có một triết lý rất mạnh: claiming work is complete without verification is dishonesty, not efficiency. Nó buộc asset trước khi khẳng định phải xác định lệnh kiểm chứng nào chứng minh được claim, chạy lệnh đó, đọc output đầy đủ, kiểm exit code và chỉ sau đó mới được kết luận. Repo cũng chỉ ra các red flags như dùng từ “should”, “probably”, tin báo cáo từ asset khác, hoặc dựa vào kiểm tra một phần. 
Đây là lớp quan trọng nhất nếu muốn đưa AI vào công việc thật.
Với cá nhân, cổng kiểm chứng có thể là:
• bài viết đã đúng brief chưa
• số liệu đã khớp nguồn chưa
• proposal đã đủ phần bắt buộc chưa
• code đã chạy test/build chưa
• bản dịch đã đúng thuật ngữ chưa
Với doanh nghiệp, verification phải được thiết kế thành gate vận hành:
• đầu ra có đầy đủ dữ kiện bắt buộc không
• có vi phạm policy không
• có trùng với nguồn chuẩn không
• confidence ở mức nào
• có cần human approval không
• có log lại được để audit không
Điểm mấu chốt là thế này: AI mạnh nhất ở việc sinh phương án. Nhưng giá trị thật trong doanh nghiệp đến từ cơ chế sàng lọc phương án.
Nói cách khác, doanh nghiệp không nên chỉ hỏi “asset tạo ra được gì”, mà phải hỏi “asset được phép công bố điều gì sau khi đã qua cổng nào”.
Nếu không có verification gate, AI chỉ tạo cảm giác hiệu quả.
Nếu có verification gate, AI mới tạo ra năng suất có thể tin cậy.
6. Lớp Tiến hóa: phiên làm việc nào cũng phải để lại tài sản
Repo có skill session-end để wrap-up cuối phiên: ghi bài học nếu đủ tính tái sử dụng, cập nhật today.md, làm mới goals. md, projects. md, cập nhật active-tasks. json. Tiêu chí ghi nhận: phải có khả năng tái dùng, trái trực giác, hoặc tiêu tốn chi phí xử lý đủ lớn. 
Từ đó, ta có một nguyên tắc rất đắt giá: mỗi phiên làm việc với AI phải để lại tài sản cho phiên sau.
Nếu hôm nay asset làm xong việc mà ngày mai lại không nhớ gì, thì hệ thống đó chưa tiến hóa. Nó chỉ đang “tiêu hao token để tạo output”.
Với cá nhân, tài sản sau mỗi phiên có thể là:
• một lesson learned
• một template prompt tốt hơn
• một checklist mới
• một skill mới
• một rule mới để tránh lỗi cũ
• một pattern đã được chuẩn hóa
Với doanh nghiệp, tài sản sau mỗi phiên nên được tích lũy ở mức hệ thống:
• case library
• rule library
• exception handling
• escalation pattern
• updated SOP
• failure mode catalog
• benchmark response mẫu
Một tổ chức biết cách học từ asset sẽ ngày càng mạnh.
Một tổ chức chỉ dùng asset để trả lời nhanh sẽ mãi dậm chân ở mức tool usage.
Nếu bạn muốn áp dụng cho công việc cá nhân trước, đừng làm quá lớn. Trước hết, viết một file “luật làm việc cá nhân” rất ngắn: AI của tôi phải ưu tiên điều gì, không được nói gì, khi nào phải verify, khi nào phải ghi nhớ. Sau đó tạo bộ nhớ tối thiểu gồm 4 file: today, projects, patterns, active-tasks.
Tiếp theo, tạo 5–7 skills quan trọng nhất cho công việc của mình. Ví dụ với người làm đào tạo, tư vấn và xây giải pháp AI như anh, bộ skill ban đầu có thể là:
• bóc tách yêu cầu khách hàng
• lên outline chương trình đào tạo
• viết proposal chuẩn enterprise
• nghiên cứu đối thủ / ngành
• QA nội dung trước khi gửi
• wrap-up sau mỗi buổi làm việc
• ghi lại lesson learned và pattern
Cuối cùng mới đến lớp asset. Lúc này asset chỉ là lớp vỏ điều phối trên các skill đó. Ví dụ:
• asset tư vấn chiến lược
• asset thiết kế chương trình đào tạo
• asset biên tập tài liệu
• asset review chất lượng
• asset quản lý tri thức dự án
Cách làm này giúp bạn không sa vào bẫy “xây asset hoành tráng nhưng rỗng ruột”.
Ta xây từ skill trước, rồi mới dựng asset.
Với doanh nghiệp, sai lầm phổ biến nhất là bắt đầu bằng công nghệ: dùng model nào, dùng nền tảng nào, tích hợp RAG ra sao?
câu hỏi đầu tiên phải là: công việc nào cần AI tham gia, với luật nào, dữ liệu nào, cổng kiểm chứng nào và bộ nhớ nào?
Tôi gợi ý một lộ trình 4 bước:
Bước 1: Chọn 1 luồng công việc thật cụ thể
Không chọn “ứng dụng AI cho phòng marketing”. Quá rộng.
Hãy chọn “AI hỗ trợ viết và QA proposal”, hoặc “AI hỗ trợ phân loại ticket và draft phản hồi”, hoặc “AI hỗ trợ tổng hợp báo cáo tuần từ nhiều nguồn”.
Bước 2: Vẽ 6 lớp cho đúng luồng đó
Luật là gì.
Bộ nhớ ở đâu.
Skill nào cần có.
asset nào chịu trách nhiệm.
Cổng kiểm chứng nào bắt buộc.
Sau mỗi phiên, học được gì để cập nhật lại hệ thống.
Bước 3: Chỉ đo 3 loại chỉ số
Tốc độ nhanh hơn bao nhiêu.
Chất lượng ổn định hơn bao nhiêu.
Tỷ lệ cần con người sửa lại giảm bao nhiêu.
Bước 4: Chuẩn hóa thành module để nhân rộng
Khi một workflow đã chạy tốt, đừng copy nguyên prompt đi khắp nơi.
Hãy chuẩn hóa thành skill, policy, memory route và verification gate.
Lúc đó doanh nghiệp mới thật sự có “nền tảng asset”, chứ không chỉ có “một vài chatbot hay”.
Muốn xây skills/asset hiệu quả, đừng bắt đầu từ prompt. Hãy bắt đầu từ kỷ luật vận hành.
Prompt chỉ là lời gọi.
Skill là đơn vị năng lực.
asset là vai trò.
Memory là lịch sử.
Verification là cổng niềm tin.
Và evolution là thứ quyết định hệ thống đó có ngày càng thông minh hơn hay không.
Một cá nhân áp dụng đúng framework này sẽ có một hệ AI ngày càng “hiểu cách mình làm việc”.
Một doanh nghiệp áp dụng đúng framework này sẽ không chỉ có chatbot, mà sẽ dần hình thành một lực lượng lao động số có vai trò, có trí nhớ và có kỷ luật làm việc.