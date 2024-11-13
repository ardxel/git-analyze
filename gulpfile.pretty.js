import gulp from 'gulp';
import fs from 'fs';
import path from 'path';
import prettier from 'gulp-prettier';

const configPath = path.relative(process.cwd(), '.prettierrc');
const prettierConfig = JSON.parse(fs.readFileSync(configPath, { encoding: 'utf8' }));

const pretty = () => {
  return gulp.src(['./gulpfile.bundle.js']).pipe(prettier(prettierConfig));
};

export default pretty;
