import React, { useState } from 'react';
import LoginModal from './LoginModal';

const LoginModalButton = () => {
    const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);

const handleOpenModal = () => {
    setIsLoginModalOpen(true);
};

return (
    <div>
    <button onClick={handleOpenModal}>Login</button>
    {isLoginModalOpen && <LoginModal onClose={() => setIsLoginModalOpen(false)} />}
    </div>
);
};

export default LoginModalButton;