import React, { useState } from 'react';
import { Users } from 'lucide-react';
import { useUsers } from './hooks/useUsers';
import { useAuth } from '../../context/AuthContext';

import UserToolbar from './components/UserToolbar';
import UserTable from './components/UserTable';
import UserFormModal from './components/UserFormModal';
import UserDeleteModal from './components/UserDeleteModal';

const UserManagement = () => {
    const { user: currentUser } = useAuth();
    const canManageUsers = currentUser?.role === 'ADMIN' || currentUser?.can_manage_users;

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
        <div className="p-6 text-slate-200 max-w-[1600px] mx-auto pb-24">
            {/* Tiêu đề trang */}
            <div className="flex items-center gap-3 mb-8">
                <div className="p-3 bg-emerald-500/20 rounded-xl text-emerald-400">
                    <Users size={24} />
                </div>
                <div>
                    <h1 className="text-2xl font-black text-white">Quản lý Nhân sự</h1>
                    <p className="text-sm text-slate-500 mt-1">Phân quyền, quản lý lịch trực và theo dõi đội ngũ SOC.</p>
                </div>
            </div>

            {/* Thanh công cụ */}
            <UserToolbar 
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                onAddUser={handleOpenAddModal}
                canManageUsers={canManageUsers}
            />

            {/* Bảng danh sách Nhân sự */}
            <UserTable 
                users={currentUsers}
                onEdit={handleOpenEditModal}
                onDelete={handleOpenDeleteModal}
                canManageUsers={canManageUsers}
            />

            {/* Phân trang */}
            {totalPages > 1 && (
                <div className="mt-4 flex justify-between items-center text-xs text-slate-500">
                    <span>Hiển thị {indexOfFirstItem + 1}-{Math.min(indexOfLastItem, filteredUsers.length)} / {filteredUsers.length}</span>
                    <div className="flex gap-2">
                        <button onClick={() => setCurrentPage(p => Math.max(p - 1, 1))} disabled={currentPage === 1} className="px-3 py-1.5 bg-slate-800 rounded hover:text-white disabled:opacity-50">Trước</button>
                        <button onClick={() => setCurrentPage(p => Math.min(p + 1, totalPages))} disabled={currentPage === totalPages} className="px-3 py-1.5 bg-slate-800 rounded hover:text-white disabled:opacity-50">Sau</button>
                    </div>
                </div>
            )}

        
            
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