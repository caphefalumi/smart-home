\<template>
  <div class="chart-container">
  <div v-if="loading" class="chart-placeholder">Loading trends...</div>
    <div v-else-if="!labels.length" class="chart-placeholder">No trend data yet</div>
    <Line v-else :data="chartData" :options="chartOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Line } from 'vue-chartjs';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler);

interface DatasetConfig {
  label: string;
  data: number[];
  borderColor?: string;
  backgroundColor?: string;
  tension?: number;
  fill?: boolean;
}

const props = defineProps<{
  labels: string[];
  datasets: DatasetConfig[];
  loading?: boolean;
}>();

const chartData = computed(() => ({
  labels: props.labels,
  datasets: props.datasets
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index' as const,
    intersect: false
  },
  plugins: {
    legend: {
      display: true,
      position: 'top' as const
    },
    tooltip: {
      mode: 'index' as const,
      intersect: false
    },
    title: {
      display: false
    }
  },
  scales: {
    x: {
      ticks: {
        autoSkip: true,
        maxTicksLimit: 8
      }
    },
    y: {
      beginAtZero: true
    }
  }
};
</script>

<style scoped>
.chart-container {
  height: 260px;
  position: relative;
}

.chart-placeholder {
  align-items: center;
  color: #666;
  display: flex;
  height: 100%;
  justify-content: center;
}
</style>
