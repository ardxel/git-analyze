import gulp from 'gulp';
import fs from 'fs';
import path from 'path';
import prettier from 'gulp-prettier';

const filesGlob = ['./gulpfile*', './assets/js/*.js', './assets/templates/*.html'];
const configPath = path.relative(process.cwd(), '.prettierrc');
const prettierConfig = JSON.parse(fs.readFileSync(configPath, { encoding: 'utf8' }));

const pretty = () => {
  return gulp.src(filesGlob).pipe(prettier(prettierConfig));
};

export default pretty;
