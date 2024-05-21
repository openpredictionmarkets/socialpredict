export const renderPersonalLinks = () => {
    const linkKeys = ['personalink1', 'personalink2', 'personalink3', 'personalink4'];
    return linkKeys.map(key => {
        const link = userData[key];
        return link ? (
            <div key={key} className='nav-link text-info-blue hover:text-blue-800'>
                <a
                    href={link}
                    target='_blank'
                    rel='noopener noreferrer'
                >
                    {link}
                </a>
            </div>
        ) : null;
    });
};
