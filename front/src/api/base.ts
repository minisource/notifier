export abstract class BaseApi {
  protected basePath: string;
  constructor(basePath: string) {
    this.basePath = basePath;
  }
  protected url(path: string = ''): string {
    return `${this.basePath}${path}`;
  }
}
