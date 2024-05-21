import React from 'react';

const ProfileModal = ({ isOpen, onClose, children, title }) => {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
            <div className="relative bg-blue-900 p-6 rounded-lg text-white max-w-4xl mx-auto">
                <h2 className="text-xl mb-4">{title}</h2>
                {children}
                <button onClick={onClose} className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white">
                    ✕
                </button>
            </div>
        </div>
    );
};

export default ProfileModal;
