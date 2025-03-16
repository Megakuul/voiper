<script>
  let {
    color,
    ...restProps
  } = $props();
</script>

<svg viewBox="0 0 3800 4600" fill="none" xmlns="http://www.w3.org/2000/svg" {...restProps}>
  <!-- phone -->
  <path id="phone" d="M635.706 475.546C638.091 445.649 657.947 421.074 686.249 411.149L1621.22 83.2749C1662.65 68.7469 1707.82 91.7432 1720.45 133.79L2017.3 1122.26C2029.65 1163.38 2005.98 1206.66 1964.7 1218.45L1413.99 1375.76C1376.29 1386.53 1352.76 1423.78 1359.91 1462.33C1393.77 1644.95 1496.34 2156.14 1638.64 2482.63C1769.24 2782.3 2041.58 3174.75 2141.34 3314.41C2163.02 3344.76 2203.35 3354.47 2236.79 3337.95L2747.38 3085.59C2785.6 3066.7 2831.9 3082.46 2850.65 3120.76L3300.2 4038.84C3320.02 4079.34 3300.97 4128.13 3258.95 4144.47L2398.05 4479.23C2372.57 4489.14 2344.32 4485.23 2323.16 4467.91C2131.81 4311.24 1180.52 3508.22 861.721 2775.81C533.048 2020.71 614.7 738.84 635.706 475.546Z" stroke="{color}"/>
</svg>

<style>
  #phone {
    stroke: white;
    stroke-width: 80px;
    stroke-opacity: 0.6;
    stroke-linecap: round;
    stroke-dasharray: 13000;
    stroke-dashoffset: 0;

    animation: animate-phone 4s ease-in-out infinite;
  }

  @keyframes animate-phone {
    0%, 10% {
      stroke-dashoffset: 13000;
    }
    50% {
      stroke-dashoffset: 0;
    }
    90%, 100% {
      stroke-dashoffset: 13000;
    }
  }
</style>