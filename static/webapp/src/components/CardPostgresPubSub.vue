<template>
  <div class="pt-8 px-4 pb-12">
    <h1 class="text-gray-700 text-2xl font-light text-center mb-4">POC - Postgres PUB/SUB</h1>

    <div class="grid grid-cols-2-1-1 gap-x-1">
      <div class="border border-white rounded-md p-4">
        <h2 class="text-gray-600 text-xl font-light text-center mb-2">Criar evento</h2>
        <div class="flex gap-x-1">
          <input v-model="event" type="text" placeholder="Nome do evento..." class="input input--blue">
          <button @click="clickCreateEvent()" class="btn btn--blue">Criar</button>
        </div>
      </div>

      <CardEvent color="yellow" title="Eventos em andamento" :events="eventsInProgress" />
      <CardEvent color="green" title="Eventos concluídos" :events="completedEvents" />
    </div>
  </div>
</template>

<script>
import CardEvent from "@/components/CardEvent";
import store from "@/store";
import api from "@/services/api";

export default {
  name: 'CardPostgresPubSub',

  components: {
    CardEvent,
  },

  data: () => ({
    sharedState: store.state,
    event: '',
  }),

  computed: {
    eventsInProgress: function() { return this.sharedState.eventsInProgress },
    completedEvents: function() { return this.sharedState.completedEvents },
  },

  methods: {
    async clickCreateEvent() {
      try {
        const response = await api.saveNewEvent({ event: this.event });
        console.debug("ok! Evento criado!", { response });
        this.event = '';
      } catch (error) {
        console.debug('Erro ao criar evento', { error });
      }
    }
  }
}
</script>
