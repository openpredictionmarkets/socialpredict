const MenuItems = ({ isLoggedIn, onLogout }) => {
    const handleLogoutClick = () => {
    if (onLogout) {
      onLogout(); // Call the onLogout function
    }
    };
}