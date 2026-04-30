import React, { useState } from 'react';
import { Users } from 'lucide-react';
import { useUsers } from './hooks/useUsers';
import { useAuth } from '../../context/AuthContext';

import UserToolbar from './components/UserToolbar';
import UserTable from './components/UserTable';
import UserFormModal from './components/UserFormModal';
import UserDeleteModal from './components/UserDeleteModal';
import GroupManagementModal from './components/GroupManagementModal'; // Component mới

const UserManagement = () => {
    const { user: currentUser } = useAuth();
    const canManageUsers = currentUser?.permissions?.user_manage === true;
    const canManageGroups = currentUser?.permissions?.group_manage === true;

    const {
        currentUsers, 
        groups, // Lấy thêm dữ liệu group từ hook
        fetchGroups,
        filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser, updateUser 
    } = useUsers();

    // State điều khiển các Modal
    const [showFormModal, setShowFormModal] = useState(false);
    const [showGroupModal, setShowGroupModal] = useState(false);
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
            
            <h1 className="text-2xl font-bold text-white flex items-center gap-3 mb-1">
                <Users size={28} className="text-indigo-400"/>
                Quản lý Nhân sự & Phòng ban
            </h1>
            <p className="text-sm text-slate-400 mb-6">Phân quyền, quản lý lịch trực và theo dõi đội ngũ SOC.</p>

            <UserToolbar 
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                onAddUser={handleOpenAddModal}
                canManageUsers={canManageUsers}
                onManageGroups={() => setShowGroupModal(true)}
                canManageGroups={canManageGroups}
            />

            <div className="flex-1 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl flex flex-col overflow-hidden">
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
                groups={groups} // Truyền danh sách group vào form
            />

            <UserDeleteModal 
                isOpen={deleteModalOpen}
                onClose={() => setDeleteModalOpen(false)}
                user={userToDelete}
                onConfirm={handleDeleteConfirm}
                isLoading={isLoading}
            />

            {/* Modal Quản lý phòng ban */}
            <GroupManagementModal 
                isOpen={showGroupModal}
                onClose={() => {
                    setShowGroupModal(false);
                    fetchGroups(); // Refresh lại danh sách sau khi đóng
                }}
                currentUser={currentUser}
            />
        </div>
    );
};

export default UserManagement;