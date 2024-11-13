import gulp from 'gulp';
import postcss from 'gulp-postcss';
import autoprefixer from 'autoprefixer';
import cssnano from 'cssnano';
import terser from 'gulp-terser';
import babel from 'gulp-babel';
import path from 'path';

const assetsPath = 'assets';
const distPath = 'dist';

const paths = {
  styles: {
    src: path.join(assetsPath, 'css/*css'),
    dest: path.join(distPath, 'css'),
  },
  scripts: {
    src: path.join(assetsPath, 'js/*.js'),
    dest: path.join(distPath, 'js'),
  },
  images: {
    src: path.join(assetsPath, 'img/*'),
    dest: path.join(distPath, 'img'),
  },
  html: {
    src: path.join(assetsPath, 'templates/*.html'),
    dest: path.join(distPath, 'templates'),
  },
};
const styles = () => {
  return gulp
    .src(paths.styles.src)
    .pipe(postcss([autoprefixer(), cssnano()]))
    .pipe(paths.styles.dest);
};

const scripts = () => {
  return gulp
    .src(paths.scripts.src)
    .pipe(babel({ presets: ['@babel/preset-env'] }))
    .pipe(terser({ mangle: true, toplevel: true }))
    .pipe(paths.scripts.dest);
};

const images = () => {
  return gulp.src(paths.images.src, { encoding: false }).pipe(paths.images.dest);
};

const html = () => {
  return gulp.src(paths.html.src).pipe(paths.html.dest);
};

const build = gulp.series(gulp.parallel(styles, scripts, images, html));

export default build;
