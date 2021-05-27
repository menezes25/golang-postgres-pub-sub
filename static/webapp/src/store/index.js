export default {
  debug: true,

  state: {
    eventsInProgress: [],
    completedEvents: [],
  },

  addEventsInProgress (newValue) {
    if (this.debug) console.log('addEventsInProgress triggered with', newValue);
    this.state.eventsInProgress.push(newValue);
  },

  addCompletedEvents (newValue) {
    if (this.debug) console.log('addCompletedEvents triggered with', newValue);
    this.state.completedEvents.push(newValue);
  },
};
