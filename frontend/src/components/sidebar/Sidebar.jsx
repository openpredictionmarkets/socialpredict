import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../../helpers/AuthContent'; // Ensure correct import path
import LoginModalButton from '../modals/login/LoginModalClick';
import { AboutSVG,
    AdminGearSVG,
    ApiKeySVG,
    CoinsSVG,
    CreateSVG,
    HomeSVG,
    LockPasswordSVG,
    LogoutSVG,
    MarketsSVG,
    MenuGrowSVG,
    MenuShrinkSVG,
    NotificationsSVG,
    PollsSVG,
    ProfileSVG }  from '../../assets/components/SvgIcons';


const Sidebar = () => {
    // Extend useAuth hook to get changePasswordNeeded along with usertype, isLoggedIn, and logout
    const { isLoggedIn, usertype, logout, changePasswordNeeded } = useAuth();
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);

    const toggleSidebar = () => setIsSidebarOpen(!isSidebarOpen);

    useEffect(() => {
        const handleOutsideClick = (event) => {
            // Close sidebar if click outside and sidebar is open
            if (!document.getElementById('sidebar').contains(event.target) && isSidebarOpen) {
                setIsSidebarOpen(false);
            }
        };

        document.addEventListener('mousedown', handleOutsideClick);
        return () => document.removeEventListener('mousedown', handleOutsideClick);
    }, [isSidebarOpen]);

    const handleLogoutClick = () => {
        logout();
        setIsSidebarOpen(false); // Also close sidebar on logout
    };

    return (
        <>
            <aside id="sidebar" className={`fixed top-0 left-0 z-30 w-sidebar h-screen bg-gray-800 text-white transform transition-transform duration-300 ease-in-out ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'} md:translate-x-0 flex flex-col`}>
                <button onClick={toggleSidebar} className="p-4 md:hidden">
                    {isSidebarOpen ? <MenuShrinkSVG /> : <MenuGrowSVG />}
                </button>
                <div className="flex-grow overflow-y-auto px-3 py-4">
                    <ul className="space-y-8 font-medium">
                        {/* Conditional rendering based on auth state */}
                        {isLoggedIn ? (
                        <>
                        {usertype === 'ADMIN' ? (
                            // Links for ADMIN
                            <>
                                <li>
                                    <Link to="/admin" className="sidebar-link">
                                        <AdminGearSVG className="mr-2" />
                                        Dashboard
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/markets" className="sidebar-link flex items-center">
                                        <MarketsSVG className="mr-2" />
                                        Markets
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/polls" className="sidebar-link flex items-center">
                                        <PollsSVG className="mr-2" />
                                        Polls
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/about" className="sidebar-link flex items-center">
                                        <AboutSVG className="mr-2" />
                                        About
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/" onClick={handleLogoutClick} className="sidebar-link flex items-center">
                                        <LogoutSVG className="mr-2" />
                                        Logout
                                    </Link>
                                </li>
                            </>
                        ) : changePasswordNeeded ? (
                            // Limited Links for Regular Users needing password change
                            <>
                                <li>
                                    <Link to="/changepassword" className="sidebar-link">
                                        <LockPasswordSVG className="mr-2" />
                                        Change Password
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/" onClick={handleLogoutClick} className="sidebar-link flex items-center">
                                        <LogoutSVG className="mr-2" />
                                        Logout
                                    </Link>
                                </li>
                            </>
                        ) : (
                            // Full Links for Regular Users
                            <>
                                <li>
                                    <Link to="/profile" className="sidebar-link flex items-center">
                                        <ProfileSVG className="mr-2" />
                                        Profile
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/markets" className="sidebar-link flex items-center">
                                        <MarketsSVG className="mr-2" />
                                        Markets
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/polls" className="sidebar-link flex items-center">
                                        <PollsSVG className="mr-2" />
                                        Polls
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/notifications" className="sidebar-link flex items-center">
                                        <NotificationsSVG className="mr-2" />
                                        Alerts
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/create" className="sidebar-link flex items-center">
                                        <CreateSVG className="mr-2" />
                                        Create
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/about" className="sidebar-link flex items-center">
                                        <AboutSVG className="mr-2" />
                                        About
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/" onClick={handleLogoutClick} className="sidebar-link flex items-center">
                                        <LogoutSVG className="mr-2" />
                                        Logout
                                    </Link>
                                </li>
                            </>
                        )}
                    </>
                        ) : (
                            <>
                                <li>
                                    <LoginModalButton />
                                </li>
                                <li>
                                    <Link to="/markets" className="flex items-center">
                                        <MarketsSVG className="mr-2" />
                                        Markets
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/polls" className="flex items-center">
                                        <PollsSVG className="mr-2" />
                                        Polls
                                    </Link>
                                </li>
                                <li>
                                    <Link to="/about" className="flex items-center">
                                        <AboutSVG className="mr-2" />
                                        About
                                    </Link>
                                </li>
                            </>
                        )}
                    </ul>
                </div>
                <footer className="border-t border-gray-200">
                    <p className="text-sm text-center p-4">📈 Built with SocialPredict. <a href="https://github.com/openpredictionmarkets/socialpredict" target="_blank" rel="noopener noreferrer" className="hover:text-blue-700">⭐ Star Us on Github!</a></p>
                </footer>
            </aside>
            {!isSidebarOpen && (
                <div className="fixed bottom-0 left-0 right-0 z-50 bg-gray-800 text-white flex justify-around items-center p-4 md:hidden">
                    {/* Minimal icons for bottom bar when sidebar is closed */}
                        {/* Conditional rendering based on auth state */}
                        {isLoggedIn ? (
                        <>
                        {usertype === 'ADMIN' ? (
                            // Links for ADMIN
                            <>
                                <div className="flex items-center space-x-4">
                                    <div className="mr-6">
                                    <Link to="/admin"><AdminGearSVG /></Link>
                                    </div>
                                    <Link to="/markets" className="flex items-center space-x-4">
                                        <div className="mr-6">
                                            <MarketsSVG />
                                        </div>
                                    </Link>
                                </div>
                            </>
                        ) : changePasswordNeeded ? (
                            // Limited Links for Regular Users needing password change
                            <>
                                <div className="flex items-center space-x-4">
                                    <div className="mr-6">
                                    <Link to="/changepassword"><LockPasswordSVG /></Link>
                                    </div>
                                </div>
                            </>
                        ) : (
                            // Full Links for Regular Logged In Users
                            <>
                                <div className="flex items-center space-x-4">
                                    <div className="mr-6">
                                    <Link to="/profile"><ProfileSVG /></Link>
                                    </div>
                                    <Link to="/markets" className="flex items-center space-x-4">
                                        <div className="mr-6">
                                            <MarketsSVG />
                                        </div>
                                    </Link>
                                    <Link to="/create" className="flex items-center space-x-4">
                                        <div className="mr-4">
                                            <CreateSVG />
                                        </div>
                                    </Link>
                                </div>
                            </>
                        )}
                    </>
                        ) : (
                            <>
                                <div className="flex items-center space-x-4">
                                    <div className="mr-6">
                                        <LoginModalButton />
                                    </div>
                                    <Link to="/markets" className="flex items-center space-x-4">
                                        <div className="mr-6">
                                            <MarketsSVG />
                                        </div>
                                    </Link>
                                    <Link to="/about" className="flex items-center space-x-4">
                                        <div className="mr-4">
                                            <AboutSVG />
                                        </div>
                                    </Link>
                                </div>
                            </>
                        )}
                    <button onClick={toggleSidebar}>
                        <MenuGrowSVG />
                    </button>
                </div>
            )}
        </>
    );
};

export default Sidebar;