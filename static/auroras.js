// Get canvas and context
const canvas = document.getElementById('auroraCanvas');
const ctx = canvas.getContext('2d');

// Make canvas full screen
canvas.width = window.innerWidth;
canvas.height = window.innerHeight;

// Handle resize
window.addEventListener('resize', () => {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    createWaves(); // Recreate waves on resize
});

// Wave configurations
const waves = [];
const baseColors = [
    { r: 0, g: 255, b: 140 },    // Green
    { r: 80, g: 200, b: 255 },   // Blue
    { r: 120, g: 100, b: 255 },  // Purple
    { r: 200, g: 50, b: 255 },   // Pink
    { r: 0, g: 180, b: 180 }     // Teal
];

// Simplified Perlin noise implementation
const perlin = {
    // Hash lookup table
    permutation: [],

    // Initialize permutation table
    init: function () {
        this.permutation = [];
        const p = Array.from({ length: 256 }, (_, i) => i);

        // Fisher-Yates shuffle
        for (let i = p.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [p[i], p[j]] = [p[j], p[i]];
        }

        // Extend to avoid overflow
        this.permutation = p.concat(p);
    },

    // Linear interpolation
    lerp: function (a, b, t) {
        return a + t * (b - a);
    },

    // Fade function for smoother interpolation
    fade: function (t) {
        return t * t * t * (t * (t * 6 - 15) + 10);
    },

    // 1D noise
    noise1D: function (x) {
        const X = Math.floor(x) & 255;
        x -= Math.floor(x);

        const u = this.fade(x);

        const a = this.permutation[X];
        const b = this.permutation[X + 1];

        // Calculate gradients and dot products
        const gradA = 2 * (a & 1) - 1;
        const gradB = 2 * (b & 1) - 1;

        // Dot products
        const n0 = gradA * x;
        const n1 = gradB * (x - 1);

        // Interpolate
        return this.lerp(n0, n1, u) * 0.5 + 0.5;
    }
};

// Initialize Perlin noise
perlin.init();

// Color utility function - HSL to RGB
function hslToRgb(h, s, l) {
    let r, g, b;

    if (s === 0) {
        r = g = b = l; // achromatic
    } else {
        const hue2rgb = (p, q, t) => {
            if (t < 0) t += 1;
            if (t > 1) t -= 1;
            if (t < 1 / 6) return p + (q - p) * 6 * t;
            if (t < 1 / 2) return q;
            if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
            return p;
        };

        const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
        const p = 2 * l - q;

        r = hue2rgb(p, q, h + 1 / 3);
        g = hue2rgb(p, q, h);
        b = hue2rgb(p, q, h - 1 / 3);
    }

    return {
        r: Math.round(r * 255),
        g: Math.round(g * 255),
        b: Math.round(b * 255)
    };
}

// Create waves with different properties
function createWaves() {
    waves.length = 0; // Clear existing waves

    const numWaves = 4;

    // Create waves with different properties, evenly spaced vertically
    for (let i = 0; i < numWaves; i++) {
        // Position waves evenly from top to bottom
        const verticalPosition = i / (numWaves - 1);
        const baseHeight = canvas.height * (0 + verticalPosition * 0.9); // 10% from top to 90% from top

        // Random wave properties
        const wave = {
            baseY: baseHeight,
            amplitude: 15 + Math.random() * 100,
            frequency: 0.003 + Math.random() * 0.002, // Lower frequency for slower waves
            speed: 0.05 + Math.random() * 0.15, // Slower speed
            phase: Math.random() * Math.PI * 2,
            // Perlin noise settings
            perlin: {
                scale: 0.0008 + Math.random() * 0.005, // Lower scale for slower variation
                speed: 0.03 + Math.random() * 0.08, // Slower perlin movement
                amplitude: 15 + Math.random() * 40,
                offset: Math.random() * 1000
            },
            // Wave height (thickness)
            height: 400 + Math.random() * 100,
            // Color settings
            baseColorIndex: Math.floor(Math.random() * baseColors.length),
            hueSpeed: 0.02 + Math.random() * 0.05, // Speed of hue shift
            hueOffset: Math.random() * 360 // Starting hue offset
        };

        waves.push(wave);
    }
}

// Calculate wave Y position with Perlin noise
function getWaveY(wave, x, time) {
    // Primary sine wave
    const primary = Math.sin(
        x * wave.frequency +
        time * wave.speed +
        wave.phase
    ) * wave.amplitude;

    // Perlin noise for irregularity
    const noiseInput = x * wave.perlin.scale + time * wave.perlin.speed + wave.perlin.offset;
    const noiseValue = perlin.noise1D(noiseInput) * 2 - 1; // -1 to 1
    const perlinOffset = noiseValue * wave.perlin.amplitude;

    // Add a slow vertical drift
    const drift = Math.sin(time * 0.05) * 10;

    // Combine all components
    return wave.baseY + primary + perlinOffset + drift;
}

// Get wave color with horizontal hue shift
function getWaveColor(wave, x, time) {
    // Get base color
    const baseColor = baseColors[wave.baseColorIndex];

    // Calculate hue shift based on position and time
    const hueShift = (x / canvas.width * 60) + (time * wave.hueSpeed * 20) + wave.hueOffset;

    // Convert HSL to RGB with the shifted hue
    // We'll use the base color to determine approximate HSL values
    const r = baseColor.r / 255;
    const g = baseColor.g / 255;
    const b = baseColor.b / 255;

    const max = Math.max(r, g, b);
    const min = Math.min(r, g, b);

    // Calculate approximate lightness
    const lightness = (max + min) / 2;

    // Calculate approximate saturation
    let saturation;
    if (max === min) {
        saturation = 0;
    } else {
        saturation = lightness > 0.5 ?
            (max - min) / (2 - max - min) :
            (max - min) / (max + min);
    }

    // Use these values with the shifted hue
    const newHue = ((hueShift % 360) / 360);
    const newColor = hslToRgb(newHue, saturation, lightness);

    return `rgba(${newColor.r}, ${newColor.g}, ${newColor.b}, 0.4)`;
}

// Draw a single wave with horizontal hue gradient
function drawWave(wave, time) {
    // Create points for the wave path
    const points = [];
    const step = 150; // Larger step for better performance

    for (let x = -0.5 * step; x <= canvas.width + step; x += step) {
        points.push({
            x: x,
            y: getWaveY(wave, x, time),
            color: getWaveColor(wave, x, time)
        });
    }

    // Draw segments with changing colors
    for (let i = 0; i < points.length - 1; i++) {
        const gradient = ctx.createLinearGradient(
            points[i].x, 0,
            points[i + 1].x, 0
        );

        gradient.addColorStop(0, points[i].color);
        gradient.addColorStop(1, points[i + 1].color);

        // Draw this segment
        ctx.beginPath();
        ctx.moveTo(points[i].x, points[i].y);
        ctx.lineTo(points[i + 1].x, points[i + 1].y);
        ctx.lineTo(points[i + 1].x, canvas.height);
        ctx.lineTo(points[i].x, canvas.height);
        ctx.closePath();

        // Create vertical gradient for this segment
        const vertGradient = ctx.createLinearGradient(
            0, points[i].y - wave.amplitude - wave.perlin.amplitude,    // x, y for start
            0, points[i].y + wave.height                                // x, y for end
        );

        vertGradient.addColorStop(0, points[i].color);
        vertGradient.addColorStop(1, 'rgba(0, 0, 0, 0)'); // transparent


        ctx.fillStyle = vertGradient;
        ctx.fill();
    }
}

// Main animation loop
function draw() {
    // Clear the canvas with slight trail effect
    ctx.fillStyle = 'rgb(30, 20, 70)';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    const time = Date.now() / 1000;

    // Draw each wave
    for (let i = 0; i < waves.length; i++) {
        drawWave(waves[i], time);
    }

    // Continue animation
    requestAnimationFrame(draw);
}

// Initialize and start
createWaves();
draw();