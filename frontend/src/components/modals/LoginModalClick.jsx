import React, { useState } from 'react';
import LoginModal from './LoginModal';
import { useAuth } from '../../helpers/AuthContent';

const LoginModalButton = () => {
    const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);
    const { login } = useAuth();

    const handleOpenModal = () => {
        setIsLoginModalOpen(true);
    };

    return (
        <div>
            <button onClick={handleOpenModal}>Login</button>
            {isLoginModalOpen && <LoginModal isOpen={isLoginModalOpen} onClose={() => setIsLoginModalOpen(false)} onLogin={login} />}
        </div>
    );
};

export default LoginModalButton;
