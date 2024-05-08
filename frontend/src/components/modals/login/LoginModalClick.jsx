import React, { useState } from 'react';
import LoginModal from './LoginModal';
import { useAuth } from '../../../helpers/AuthContent';
import { useHistory } from 'react-router-dom';

const LoginModalButton = () => {
    const [isLoginModalOpen, setIsLoginModalOpen] = useState(false);
    const { login } = useAuth();
    const [redirectAfterLogin, setRedirectAfterLogin] = useState('/');
    const history = useHistory();

    const handleOpenModal = () => {
        setRedirectAfterLogin(history.location.pathname);
        setIsLoginModalOpen(true);
    };

    return (
        <div>
            <button onClick={handleOpenModal}>Login</button>
            {isLoginModalOpen && <LoginModal isOpen={isLoginModalOpen} onClose={() => setIsLoginModalOpen(false)} onLogin={login} redirectAfterLogin={redirectAfterLogin} />}
        </div>
    );
};

export default LoginModalButton;
