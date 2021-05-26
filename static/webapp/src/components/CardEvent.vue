<template>
  <div class="bg-gray-600 rounded-md p-1 text-white">
    <h2 class="font-light text-center">
      {{title}}
    </h2>
    <ul
      class="my-4 list-disc list-inside pl-4"
      :class="[ cardEventColorClass ]"
    >
      <li v-for="(event, idx) in events" :key="idx">
        {{event}}
      </li>
    </ul>
  </div>
</template>

<script>
// https://tailwindcss.com/docs/optimizing-for-production#writing-purgeable-html
// Uso do nome da classe literal para compatibilidade PurgeCSS
const cardEventColors = new Map()
  .set('yellow', 'list-marker-yellow')
  .set('green', 'list-marker-green');

export default {
  name: 'CardEvent',

  props: {
    title: { type: String, required: true },
    events: { type: Array, required: true },
    color: { type: String, required: true, validator: (value) => cardEventColors.has(value) },
  },

  computed: {
    cardEventColorClass() {
      return cardEventColors.get(this.color);
    },
  }
}
</script>
