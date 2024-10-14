const gulp = require("gulp");
const postcss = require("gulp-postcss");
const autoprefixer = require("autoprefixer");
const cssnano = require("cssnano");
const terser = require("gulp-terser");
const babel = require("gulp-babel");

const paths = {
  styles: {
    src: "./static/css/*.css",
    dest: "./dist/css",
  },
  scripts: {
    src: "./static/js/*.js",
    dest: "./dist/js",
  },
  images: {
    src: "./static/img/*",
    dest: "./dist/img",
  },
};

const styles = () => {
  return gulp
    .src(paths.styles.src)
    .pipe(postcss([autoprefixer(), cssnano()]))
    .pipe(gulp.dest(paths.styles.dest));
};

const scripts = () => {
  return gulp
    .src(paths.scripts.src)
    .pipe(babel({ presets: ["@babel/preset-env"] }))
    .pipe(terser({ mangle: true, toplevel: true }))
    .pipe(gulp.dest(paths.scripts.dest));
};

const images = () => {
  return gulp.src(paths.images.src).pipe(gulp.dest(paths.images.dest));
};

const build = gulp.series(gulp.parallel(styles, scripts, images));

exports.build = build;
exports.default = build;
