import React from 'react';
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
    LoginSVG,
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

    const handleLogoutClick = () => {
        logout();
    };

    return (
        <aside className="fixed top-0 left-0 z-sidebar w-sidebar h-screen flex flex-col" aria-label="Sidebar">
            <div className="flex-grow overflow-y-auto px-3 py-4 dark:bg-custom-gray-dark">
                <ul className="space-y-2 font-medium">
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
                        // Links when not logged in
                        <>
                            <li>
                                <LoginModalButton />
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
                        </>
                    )}
                </ul>
            </div>
            <div className="border-t border-gray-200 dark:bg-custom-gray-dark">
                <footer className="p-4">
                    <p className="text-sm text-center">📈 Built with SocialPredict. <a href="https://github.com/openpredictionmarkets/socialpredict" target="_blank" rel="noopener noreferrer" className="text-blue-500 hover:text-blue-700">⭐ Star Us on Github!</a></p>
                </footer>
            </div>
        </aside>
    );
};

export default Sidebar;
