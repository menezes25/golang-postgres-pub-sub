<template>
  <div id="app" class="h-full w-full pt-16 bg-gradient-to-r from-blueGrey to-cadetBlue">
    <div class="container mx-auto max-w-screen-md rounded-lg bg-gray-100 bg-opacity-50">
      <CardPostgresPubSub />
    </div>
  </div>
</template>

<script>
import CardPostgresPubSub from "@/components/CardPostgresPubSub";

export default {
  name: 'App',
  components: {
    CardPostgresPubSub,
  }
}
</script>
