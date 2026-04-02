import React, { useState } from 'react';
import { Users, Search, Tag } from 'lucide-react';
import { useUsers } from './hooks/useUsers';
import { useAuth } from '../../context/AuthContext';

import UserToolbar from './components/UserToolbar';
import UserTable from './components/UserTable';
import UserFormModal from './components/UserFormModal';
import UserDeleteModal from './components/UserDeleteModal';

const UserManagement = () => {
    const { user: currentUser } = useAuth();
    const canManageUsers = currentUser?.permissions?.user_manage === true;

    const {
        currentUsers, filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser, updateUser 
    } = useUsers();

    // State điều khiển các Modal
    const [showFormModal, setShowFormModal] = useState(false);
    const [isEditMode, setIsEditMode] = useState(false);
    const [userToEdit, setUserToEdit] = useState(null);
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [userToDelete, setUserToDelete] = useState(null);

    // --- CÁC HÀM XỬ LÝ SỰ KIỆN CHUẨN ---
    const handleOpenAddModal = () => {
        setIsEditMode(false);
        setUserToEdit(null);
        setShowFormModal(true);
    };

    const handleOpenEditModal = (user) => {
        setIsEditMode(true);
        setUserToEdit(user);
        setShowFormModal(true);
    };

    const handleOpenDeleteModal = (user) => {
        setUserToDelete(user);
        setDeleteModalOpen(true);
    };

    const handleFormSubmit = async (formData) => {
        if (isEditMode) {
            await updateUser(userToEdit.id, formData);
        } else {
            await createUser(formData);
        }
        setShowFormModal(false);
    };

    const handleDeleteConfirm = async (id) => {
        await deleteUser(id);
        setDeleteModalOpen(false);
    };

    return (
        <div className="p-6 text-slate-200 h-[calc(100vh-60px)] flex flex-col bg-[#050B14] font-sans">
            
            <div className="flex justify-between items-end mb-4 shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <Users size={24} className="text-indigo-500"/> QUẢN LÝ NHÂN SỰ
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Phân quyền, quản lý lịch trực và theo dõi đội ngũ SOC.</p>
                </div>
            </div>

            <div className="flex-1 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl flex flex-col overflow-hidden">
                <div className="p-3 border-b border-slate-800 bg-[#111827] flex justify-between items-center shrink-0">
                    <div className="relative w-72">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={14} />
                        <input type="text" placeholder="Search users..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full bg-[#050B14] border border-slate-800 text-xs text-white rounded pl-9 pr-4 py-1.5 outline-none focus:border-indigo-500 font-mono transition-colors" />
                    </div>
                    <div className="flex items-center gap-2">
                        <Tag className="text-emerald-500" size={12}/> <span className="text-[10px] font-mono font-bold text-slate-400">TOTAL: {filteredUsers.length}</span>
                    </div>
                </div>

                <UserTable 
                    users={currentUsers}
                    onEdit={handleOpenEditModal}
                    onDelete={handleOpenDeleteModal}
                    canManageUsers={canManageUsers}
                />

                {totalPages > 1 && (
                    <div className="mt-4 flex justify-between items-center text-xs text-slate-500 px-3 pb-3">
                        <span>Hiển thị {indexOfFirstItem + 1}-{Math.min(indexOfLastItem, filteredUsers.length)} / {filteredUsers.length}</span>
                        <div className="flex gap-2">
                            <button onClick={() => setCurrentPage(p => Math.max(p - 1, 1))} disabled={currentPage === 1} className="px-3 py-1.5 bg-slate-800 rounded hover:text-white disabled:opacity-50">Trước</button>
                            <button onClick={() => setCurrentPage(p => Math.min(p + 1, totalPages))} disabled={currentPage === totalPages} className="px-3 py-1.5 bg-slate-800 rounded hover:text-white disabled:opacity-50">Sau</button>
                        </div>
                    </div>
                )}
            </div>

        
            
            <UserFormModal 
                isOpen={showFormModal} 
                onClose={() => setShowFormModal(false)}
                initialData={userToEdit}
                onSubmit={handleFormSubmit}
                isLoading={isLoading}
            />

            <UserDeleteModal 
                isOpen={deleteModalOpen}
                onClose={() => setDeleteModalOpen(false)}
                user={userToDelete}
                onConfirm={handleDeleteConfirm}
                isLoading={isLoading}
            />
        </div>
    );
};

export default UserManagement;