function startMate() {
	const root = window.MateDOM.getAppRoot();

	if (!root) {
		console.error("MATE app root was not found");
		return;
	}

	console.log("MATE frontend loaded");
}

startMate();
