const formatResolutionDate = (resolutionDateTime) => {
  const now = new Date();
  const resolutionDate = new Date(resolutionDateTime);

  return resolutionDate < now ? 'Closed' : resolutionDate.toLocaleDateString();
};

export default formatResolutionDate;
