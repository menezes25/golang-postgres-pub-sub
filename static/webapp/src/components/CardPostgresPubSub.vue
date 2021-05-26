<template>
  <div class="pt-8 px-4 pb-12">
    <h1 class="text-gray-700 text-2xl font-light text-center mb-4">POC - Postgres PUB/SUB</h1>

    <div class="grid grid-cols-2-1-1 gap-x-1">
      <div class="border border-white rounded-md p-4">
        <h2 class="text-gray-600 text-xl font-light text-center mb-2">Criar evento</h2>
        <div class="flex gap-x-1">
          <input type="text" placeholder="Nome do evento..." class="border border-gray-400 p-2 rounded focus:outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-opacity-80">
          <button class="btn btn--blue">Criar</button>
        </div>
      </div>

      <div class="bg-gray-600 rounded-md p-1 text-white">
        <h2 class="font-light text-center">Eventos em andamento</h2>
        <ul class="my-4 list-disc list-inside list-marker-yellow pl-4">
          <li>demo a</li>
          <li>demo b</li>
        </ul>
      </div>

      <div class="bg-gray-600 rounded-md p-1 text-white">
        <h2 class="font-light text-center">Eventos concluídos</h2>
        <ul class="my-4 list-disc list-inside list-marker-green pl-4">
          <li>demo a</li>
          <li>demo b</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CardPostgresPubSub',
}
</script>
