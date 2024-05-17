const Tooltip = ({ content, position }) => {
  if (!content) return null;

  const style = {
    display: 'block',
    position: 'absolute',
    top: `${position.y}px`,
    left: `${position.x}px`,
    backgroundColor: 'white',
    padding: '10px',
    border: '1px solid',
    borderRadius: '5px',
    pointerEvents: 'none', // this prevents the tooltip from interfering with mouse events
    zIndex: 1000 // make sure this is above everything else
  };

  return (
    <div id="tooltip" style={style}>
      {content}
    </div>
  );
};

export default Tooltip;