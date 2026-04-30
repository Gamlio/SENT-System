// Thêm vào trong phần grid thông tin cơ bản của UserFormModal[cite: 44]
<InputRow label="Phòng ban / Nhóm" icon={Shield}>
    <select 
        value={formData.group_id || ''} 
        onChange={e => setFormData({...formData, group_id: e.target.value ? Number(e.target.value) : null})}
        className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none appearance-none"
    >
        <option value="">-- Chọn phòng ban --</option>
        {groups.map(g => (
            <option key={g.ID} value={g.ID}>{g.name}</option>
        ))}
    </select>
</InputRow>

{/* Thêm quyền Quản lý nhóm trong Ma trận đặc quyền */}
<PermissionBlock title="Quản trị Hệ thống" icon={<Settings size={16}/>} color="text-purple-400">
    <ToggleSwitch field="perm_group_manage" label="Quản lý Nhóm (Groups)" data={formData} onToggle={handleToggle}/>
    <ToggleSwitch field="perm_approval_manage" label="Duyệt yêu cầu (Maker-Checker)" data={formData} onToggle={handleToggle}/>
    <ToggleSwitch field="perm_user_manage" label="Quản lý Nhân sự" data={formData} onToggle={handleToggle} isDanger/>
</PermissionBlock>